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

package stream

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// A Message is a consumed message from a redis stream.
type Message struct {
	Stream string
	ID     string
	Values map[string]any
}

type config struct {
	group    string
	consumer string
	streams  []string // list of streams and ids, e.g. stream1 stream2 id1 id2
	count    int64
	block    time.Duration
	noAck    bool
}

// An Option modifies the config.
type Option func(*config)

// WithStream adds a stream to the consumer.
func WithStream(stream string) Option {
	return func(cfg *config) {
		cfg.streams = append(cfg.streams, stream)
	}
}

// WithCount sets the count for the config.
func WithCount(cnt int64) Option {
	return func(cfg *config) {
		cfg.count = cnt
	}
}

// WithBlock sets the block field of the config.
func WithBlock(duration time.Duration) Option {
	return func(cfg *config) {
		cfg.block = duration
	}
}

// WithNoAck sets the noAck field of the config.
func WithNoAck(noAck bool) Option {
	return func(cfg *config) {
		cfg.noAck = noAck
	}
}

// A Consumer consumes messages from a stream.
type Consumer struct {
	client  *redis.Client
	cfg     *config
	lastIDs map[string]string
}

// New creates a new consumer.
func New(client *redis.Client, group, consumer string, options ...Option) *Consumer {
	cfg := &config{
		group:    group,
		consumer: consumer,
	}
	for _, opt := range options {
		opt(cfg)
	}
	lastIDs := make(map[string]string)
	for _, stream := range cfg.streams {
		lastIDs[stream] = "0-0"
	}

	return &Consumer{
		client:  client,
		cfg:     cfg,
		lastIDs: lastIDs,
	}
}

// Read reads messages from the stream.
func (c *Consumer) Read(ctx context.Context) ([]Message, error) {
	for {
		streams := make([]string, 0, len(c.cfg.streams)*2)
		streams = append(streams, c.cfg.streams...)
		for _, stream := range c.cfg.streams {
			streams = append(streams, c.lastIDs[stream])
		}

		cmd := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    c.cfg.group,
			Consumer: c.cfg.consumer,
			Streams:  streams,
			Count:    c.cfg.count,
			Block:    c.cfg.block,
			NoAck:    c.cfg.noAck,
		})

		vals, err := cmd.Result()
		if err == redis.Nil {
			if c.cfg.block >= 0 {
				continue
			}

			return nil, nil //nolint:revive
		} else if err != nil {
			return nil, err
		}

		allLatest := true
		for _, lastID := range c.lastIDs {
			if lastID != ">" {
				allLatest = false
			}
		}

		var msgs []Message
		for _, stream := range vals {
			if len(stream.Messages) == 0 {
				c.lastIDs[stream.Stream] = ">"
			}
			for _, msg := range stream.Messages {
				msgs = append(msgs, Message{
					Stream: stream.Stream,
					ID:     msg.ID,
					Values: msg.Values,
				})
				c.lastIDs[stream.Stream] = msg.ID
			}
		}
		if len(msgs) > 0 || allLatest {
			return msgs, nil
		}
	}
}

// Ack acknowledges the messages.
func (c *Consumer) Ack(ctx context.Context, msgs ...Message) error {
	if len(msgs) == 0 {
		return nil
	}

	if len(msgs) == 1 {
		msg := msgs[0]
		return c.client.XAck(ctx, msg.Stream, c.cfg.group, msg.ID).Err()
	}

	ids := map[string][]string{}
	for _, msg := range msgs {
		ids[msg.Stream] = append(ids[msg.Stream], msg.ID)
	}

	_, err := c.client.Pipelined(ctx, func(p redis.Pipeliner) error {
		for stream, msgIDs := range ids {
			p.XAck(ctx, stream, c.cfg.group, msgIDs...)
		}
		return nil
	})
	return err
}
