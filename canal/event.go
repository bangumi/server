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
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	stream Stream,
	search search.Client,
	s3 *s3.Client,
) *eventHandler {
	return &eventHandler{
		redis:   redis,
		config:  appConfig,
		session: session,
		search:  search,
		s3:      s3,
		stream:  stream,
		log:     log.Named("eventHandler"),
	}
}

type eventHandler struct {
	closed  atomic.Bool
	config  config.AppConfig
	session session.Manager
	log     *zap.Logger
	search  search.Client
	stream  Stream
	s3      *s3.Client // optional, check nil before use
	redis   *redis.Client
}

func (e *eventHandler) start() error {
	ee := e.stream.Read(context.Background(), func(msg Msg) error {
		e.log.Debug("new message", zap.String("topic", msg.Topic), zap.String("id", msg.ID))

		err := e.onMessage(msg.Key, msg.Value)
		if err != nil {
			e.log.Error("failed to handle stream msg",
				zap.Error(err), zap.String("stream", msg.Topic), zap.String("id", msg.ID))
			return errgo.Trace(err)
		}

		return nil
	})

	return errgo.Trace(ee)
}

func (e *eventHandler) Close() error {
	e.closed.Store(true)
	e.search.Close()
	return nil
}

func (e *eventHandler) OnUserPasswordChange(ctx context.Context, id model.UserID) error {
	e.log.Info("user change password", log.User(id))

	if err := e.session.RevokeUser(ctx, id); err != nil {
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

	var p Payload
	if err := json.Unmarshal(value, &p); err != nil {
		return nil
	}

	e.log.Debug("new message", zap.String("table", p.Source.Table))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	switch p.Source.Table {
	case "chii_subject_fields":
		err = e.OnSubjectField(ctx, key, p)
	case "chii_subjects":
		err = e.OnSubject(ctx, key, p)
	case "chii_members":
		err = e.OnUserChange(ctx, key, p)
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

type Payload struct {
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
	Source source          `json:"source"`
	Op     string          `json:"op"`
}

type source struct {
	Table string `json:"table"`
}
