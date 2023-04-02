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

package canal

import (
	"context"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/canal/stream"
	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/pkg/logger"
)

func newRedisStream(cfg config.AppConfig, redisClient *redis.Client) (Stream, error) {
	var ch = make(chan Msg, 1)

	reader := stream.New(
		redisClient,
		groupID,
		"canal",
		lo.Map(cfg.Canal.Topics, func(item string, _ int) stream.Option {
			return stream.WithStream(item)
		})...,
	)

	r := &redisStream{
		log:    logger.Named("canal.search.stream"),
		ch:     ch,
		cfg:    cfg,
		redis:  redisClient,
		reader: reader,
	}

	for _, s := range r.cfg.Canal.Topics {
		var infos, err = r.redis.XInfoGroups(context.Background(), s).Result()
		if err != nil {
			if err.Error() != "ERR no such key" {
				return nil, errgo.Trace(err)
			}
		}

		groups := lo.SliceToMap(infos, func(item redis.XInfoGroup) (string, bool) {
			return item.Name, true
		})

		if !groups[groupID] {
			err := r.redis.XGroupCreateMkStream(context.Background(), s, groupID, "$").Err()
			if err != nil {
				return nil, errgo.Trace(err)
			}
		}
	}

	return r, nil
}

type redisStream struct {
	log    *zap.Logger
	redis  *redis.Client
	ch     chan Msg
	cfg    config.AppConfig
	closed atomic.Bool
	reader *stream.Consumer
}

func (r *redisStream) Read(ctx context.Context, onMessage func(msg Msg) error) error {
	for {
		if r.closed.Load() {
			return nil
		}

		rr, err := r.reader.Read(ctx)
		if err != nil {
			r.log.Error("failed to read new messages", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		for _, msg := range rr {
			r.log.Debug("new message", zap.String("id", msg.ID), zap.String("s", msg.Stream))

			rawKey := msg.Values["key"]
			rawValue := msg.Values["value"]

			if rawKey == nil || rawValue == nil {
				_ = r.reader.Ack(context.Background(), msg)
				continue
			}

			value, ok := rawValue.(string)
			if !ok {
				r.log.Error("failed to handle event", zap.String("id", msg.ID),
					zap.String("value-type", reflect.TypeOf(rawKey).String()))
				_ = r.reader.Ack(context.Background(), msg)
				continue
			}

			key, ok := rawKey.(string)
			if !ok {
				r.log.Error("failed to handle event", zap.String("id", msg.ID),
					zap.String("key-type", reflect.TypeOf(rawKey).String()))
				_ = r.reader.Ack(context.Background(), msg)
				continue
			}

			if err := onMessage(Msg{ID: msg.ID, Stream: msg.Stream, Key: []byte(key), Value: []byte(value)}); err != nil {
				if e := r.reader.Ack(ctx, msg); e != nil {
					return errgo.Trace(err)
				}
				return errgo.Trace(err)
			}
		}
	}
}

func (r *redisStream) Close() error {
	r.closed.Store(true)
	return nil
}

func (r *redisStream) Ack(ctx context.Context, msg Msg) error {
	return r.reader.Ack(ctx, stream.Message{Stream: msg.Stream, ID: msg.ID})
}
