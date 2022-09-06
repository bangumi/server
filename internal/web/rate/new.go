// SPDX-License-Identifier: AGPL-3.0-only
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>

package rate

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/web/rate/action"
)

const keyPrefix = "chii:rate:" + config.Version
const defaultAllowPerHour = 5

//go:embed allow.lua
var allowLua string

var allowScript = redis.NewScript(allowLua) //nolint:gochecknoglobals

type Manager interface {
	// Login 检查是登录限流。
	Login(ctx context.Context, ip string) (allowed bool, remain int, err error)

	AllowAction(
		ctx context.Context,
		u model.UserID,
		action action.Action,
		limit Limit,
	) (allowed bool, remain int, err error)
	// Reset 登录成功时应该重置计数。
	Reset(ctx context.Context, ip string) error
}

const RedisRateKeyPrefix = "chii:login:rate:"
const RedisBanKeyPrefix = "chii:ban:"

func New(rdb *redis.Client) Manager {
	return manager{
		r: rdb,
	}
}

type manager struct {
	r *redis.Client
}

func (m manager) AllowAction(
	ctx context.Context,
	u model.UserID,
	action action.Action,
	limit Limit,
) (bool, int, error) {
	rateKey := fmt.Sprintf("%s:%d:%d", keyPrefix, action, u)
	banKey := fmt.Sprintf("%s:ban:%d:%d", keyPrefix, action, u)

	res, err := m.allow(ctx, rateKey, banKey, limit)
	if err != nil {
		return false, 0, errgo.Wrap(err, "Limiter.Allow")
	}

	return res.Allowed > 0, res.Remaining, nil
}

func (m manager) Login(ctx context.Context, ip string) (bool, int, error) {
	var rateKey = RedisRateKeyPrefix + ip
	var banKey = RedisBanKeyPrefix + ip

	res, err := m.allow(ctx, rateKey, banKey, PerHour(defaultAllowPerHour))
	if err != nil {
		return false, 0, errgo.Wrap(err, "Limiter.Allow")
	}

	return res.Allowed > 0, res.Remaining, nil
}

func (m manager) Reset(ctx context.Context, ip string) error {
	err := m.r.Del(ctx, RedisRateKeyPrefix+ip, RedisBanKeyPrefix+ip).Err()

	return errgo.Wrap(err, "Limiter.Allow")
}

// AllowN reports whether n events may happen at time now.
func (m manager) allow(
	ctx context.Context,
	rateKey string,
	banKey string,
	limit Limit,
) (Result, error) {
	now := time.Now()
	var keys = []string{rateKey, banKey}
	scriptValues := []any{
		limit.Burst, limit.Rate, limit.Period.Seconds(), now.Unix(), now.Nanosecond() / 1000, gtime.OneWeekSec,
	}
	v, err := allowScript.Run(ctx, m.r, keys, scriptValues...).Result()
	if err != nil {
		return Result{}, errgo.Wrap(err, "luaScript.Run")
	}

	values, ok := v.([]any)
	if !ok {
		panic("failed to cast redis lua result type")
	}

	retryAfter, err := strconv.ParseFloat(values[2].(string), 64)
	if err != nil {
		return Result{}, errgo.Wrap(err, "strconv.ParseFloat")
	}

	resetAfter, err := strconv.ParseFloat(values[3].(string), 64)
	if err != nil {
		return Result{}, errgo.Wrap(err, "strconv.ParseFloat")
	}

	allowed, ok := values[0].(int64)
	if !ok {
		panic("can't convert redis result 'allowed' to int64")
	}

	remaining, ok := values[1].(int64)
	if !ok {
		panic("can't convert redis result 'remaining' to int64")
	}

	return Result{
		Limit:      limit,
		Allowed:    int(allowed),
		Remaining:  int(remaining),
		RetryAfter: dur(retryAfter),
		ResetAfter: dur(resetAfter),
	}, nil
}

type Limit struct {
	Rate   int
	Burst  int
	Period time.Duration
}

func (l Limit) String() string {
	return fmt.Sprintf("%d req/%s (burst %d)", l.Rate, fmtDur(l.Period), l.Burst)
}

func (l Limit) IsZero() bool {
	return l == Limit{}
}

func fmtDur(d time.Duration) string {
	switch d { //nolint:exhaustive
	case time.Second:
		return "s"
	case time.Minute:
		return "m"
	case time.Hour:
		return "h"
	}
	return d.String()
}

func PerSecond(rate int) Limit {
	return Limit{
		Rate:   rate,
		Period: time.Second,
		Burst:  rate,
	}
}

func PerMinute(rate int) Limit {
	return Limit{
		Rate:   rate,
		Period: time.Minute,
		Burst:  rate,
	}
}

func PerHour(rate int) Limit {
	return Limit{
		Rate:   rate,
		Period: time.Hour,
		Burst:  rate,
	}
}

func dur(f float64) time.Duration {
	if f == -1 {
		return -1
	}
	return time.Duration(f * float64(time.Second))
}

type Result struct {
	// Limit is the limit that was used to obtain this result.
	Limit Limit

	// Allowed is the number of events that may happen at time now.
	Allowed int

	// Remaining is the maximum number of requests that could be
	// permitted instantaneously for this key given the current
	// state. For example, if a rate limiter allows 10 requests per
	// second and has already received 6 requests for this key this
	// second, Remaining would be 4.
	Remaining int

	// RetryAfter is the time until the next request will be permitted.
	// It should be -1 unless the rate limit has been exceeded.
	RetryAfter time.Duration

	// ResetAfter is the time until the RateLimiter returns to its
	// initial state for a given key. For example, if a rate limiter
	// manages requests per second and received one request 200ms ago,
	// Reset would return 800ms. You can also think of this as the time
	// until Limit and Remaining will be equal.
	ResetAfter time.Duration
}
