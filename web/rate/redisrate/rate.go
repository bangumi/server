//nolint:goheader,forcetypeassert,wrapcheck
package redisrate

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

//go:embed allow.lua
var allowLua string

// Copyright (c) 2017 Pavel Pravosud
// https://github.com/rwz/redis-gcra/blob/master/vendor/perform_gcra_ratelimit.lua

var allowN = redis.NewScript(allowLua) //nolint:gochecknoglobals

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

// ------------------------------------------------------------------------------

// Limiter controls how frequently events are allowed to happen.
type Limiter struct {
	rdb *redis.Client
}

// NewLimiter returns a new Limiter.
func NewLimiter(rdb *redis.Client) *Limiter {
	return &Limiter{
		rdb: rdb,
	}
}

// Allow is a shortcut for AllowN(ctx, key, limit, 1).
func (l Limiter) Allow(ctx context.Context, key string, limit Limit) (Result, error) {
	return l.AllowN(ctx, key, limit, 1)
}

// AllowN reports whether n events may happen at time now.
func (l Limiter) AllowN(
	ctx context.Context,
	key string,
	limit Limit,
	n int,
) (Result, error) {
	now := time.Now()
	v, err := allowN.Run(
		ctx, l.rdb, []string{key}, limit.Burst, limit.Rate, limit.Period.Seconds(), n, now.Unix(), now.Nanosecond()/1000,
	).Result()
	if err != nil {
		return Result{}, err
	}

	values := v.([]interface{})

	retryAfter, err := strconv.ParseFloat(values[2].(string), 64)
	if err != nil {
		return Result{}, err
	}

	resetAfter, err := strconv.ParseFloat(values[3].(string), 64)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Limit:      limit,
		Allowed:    int(values[0].(int64)),
		Remaining:  int(values[1].(int64)),
		RetryAfter: dur(retryAfter),
		ResetAfter: dur(resetAfter),
	}, nil
}

// Reset gets a key and reset all limitations and previous usages.
func (l *Limiter) Reset(ctx context.Context, key string) error {
	return l.rdb.Del(ctx, key).Err()
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
