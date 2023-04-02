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
	"errors"
	"io"
	"strconv"
	"sync/atomic"

	"github.com/segmentio/kafka-go"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/utils"
)

func newKafkaStream(cfg config.AppConfig) Stream {
	logger.Info("new kafka stream broker")
	k := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{cfg.Canal.KafkaBroker},
		GroupID:     groupID,
		GroupTopics: cfg.Canal.Topics,
	})

	var ch = make(chan Msg, 1)

	return &kafkaStream{
		log: logger.Named("canal.search.stream.kafka"),
		k:   k,
		ch:  ch,
	}
}

type kafkaStream struct {
	log    *zap.Logger
	k      *kafka.Reader
	ch     chan Msg
	closed atomic.Bool
}

func (s *kafkaStream) Read(ctx context.Context, onMessage func(msg Msg) error) error {
	for {
		if s.closed.Load() {
			return nil
		}

		msg, err := s.k.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) || utils.IsNetworkError(err) {
				s.log.Error("network error", zap.Error(err))
				continue
			}

			s.log.Error("error fetching msg", zap.Error(err))
			continue
		}

		s.log.Debug("new message", zap.String("topic", msg.Topic))

		m := Msg{
			ID:     strconv.FormatInt(msg.Offset, 10),
			Stream: msg.Topic,
			Key:    msg.Key,
			Value:  msg.Value,
		}

		if err := onMessage(m); err != nil {
			if e := s.k.CommitMessages(ctx, msg); e != nil {
				return errgo.Trace(err)
			}
			return errgo.Trace(err)
		}
	}
}

func (s *kafkaStream) Close() error {
	s.closed.Store(true)
	close(s.ch)
	return s.k.Close()
}
