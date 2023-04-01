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
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/web/session"
)

func newEventHandler(
	log *zap.Logger,
	appConfig config.AppConfig,
	session session.Manager,
	redis *redis.Client,
	search search.Client,
	s3 *minio.Client,
) *eventHandler {
	return &eventHandler{
		config:  appConfig,
		session: session,
		search:  search,
		redis:   redis,
		s3:      s3,
		log:     log.Named("eventHandler"),
	}
}

type eventHandler struct {
	closed  atomic.Bool
	config  config.AppConfig
	session session.Manager
	log     *zap.Logger
	search  search.Client
	redis   *redis.Client
	s3      *minio.Client // optional, check nil before use
}

var errNoTopic = fmt.Errorf("missing search events topic")

func (e *eventHandler) start() error {
	if len(e.config.Search.Topics) == 0 {
		return errNoTopic
	}

	s, err := newRedisStream(e.config, e.redis)
	if err != nil {
		return errgo.Trace(err)
	}

	ch, err := s.Read(context.TODO())
	if err != nil {
		return errgo.Trace(err)
	}

	for {
		msg, ok := <-ch
		if !ok {
			// chan closed
			return nil
		}
		e.log.Debug("new message", zap.String("stream", msg.Stream), zap.String("id", msg.ID))

		err = e.onMessage(msg.Key, msg.Value)
		if err != nil {
			e.log.Error("failed to handle kafka msg", zap.Error(err), zap.String("stream", msg.Stream))
			continue
		}

		_ = s.Ack(context.TODO(), msg)
	}
}

func (e *eventHandler) Close() error {
	e.closed.Store(true)
	e.search.Close()
	return nil
}

func (e *eventHandler) OnUserPasswordChange(id model.UserID) error {
	e.log.Info("user change password", log.User(id))

	if err := e.session.RevokeUser(context.Background(), id); err != nil {
		e.log.Error("failed to revoke user", log.User(id), zap.Error(err))
		return errgo.Wrap(err, "session.RevokeUser")
	}

	return nil
}

func (e *eventHandler) onMessage(key, value []byte) error {
	if len(value) == 0 {
		// fake event, just ignore
		// https://debezium.io/documentation/reference/stable/connectors/mysql.html#mysql-tombstone-events
		return nil
	}

	var k messageKey
	if err := json.Unmarshal(key, &k); err != nil {
		return nil
	}

	var v messageValue
	if err := json.Unmarshal(value, &v); err != nil {
		return nil
	}

	e.log.Debug("new message", zap.String("table", v.Payload.Source.Table))

	var err error
	switch v.Payload.Source.Table {
	case "chii_subject_fields":
		err = e.OnSubjectField(k.Payload, v.Payload)
	case "chii_subjects":
		err = e.OnSubject(k.Payload, v.Payload)
	case "chii_members":
		err = e.OnUserChange(k.Payload, v.Payload)
	}

	return err
}

const (
	opCreate   = "c"
	opDelete   = "d"
	opUpdate   = "u"
	opSnapshot = "r" // just ignore them, production debezium disable snapshot.
)

// https://debezium.io/documentation/reference/connectors/mysql.html
// Table 9. Overview of change event basic content

type messageKey struct {
	Payload json.RawMessage `json:"payload"`
}

type messageValue struct {
	Payload payload `json:"payload"`
}

type payload struct {
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
	Source source          `json:"source"`
	Op     string          `json:"op"`
}

type source struct {
	Table string `json:"table"`
}
