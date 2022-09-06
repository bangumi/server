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
	"sync/atomic"

	"github.com/goccy/go-json"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/web/session"
)

func newEventHandler(
	log *zap.Logger,
	appConfig config.AppConfig,
	session session.Manager,
	reader *kafka.Reader,
	search search.Client,
) *eventHandler {
	return &eventHandler{
		config:  appConfig,
		session: session,
		reader:  reader,
		search:  search,
		log:     log.Named("eventHandler"),
	}
}

type eventHandler struct {
	closed  atomic.Bool
	config  config.AppConfig
	session session.Manager
	log     *zap.Logger
	search  search.Client
	reader  *kafka.Reader
}

func (e *eventHandler) start() error {
	for {
		if e.closed.Load() {
			return nil
		}

		msg, err := e.reader.FetchMessage(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return errgo.Wrap(err, "read message")
			}

			e.log.Error("error fetching msg", zap.Error(err))
			continue
		}

		e.log.Debug("new message", zap.String("topic", msg.Topic))

		err = e.onMessage(msg)
		if err != nil {
			e.log.Error("failed to handle kafka msg", zap.Error(err), zap.String("topic", msg.Topic))
			continue
		}

		_ = e.reader.CommitMessages(context.Background(), msg)
	}
}

func (e *eventHandler) Close() error {
	e.closed.Store(true)
	return errgo.Wrap(e.reader.Close(), "kafka.Close")
}

func (e *eventHandler) OnUserPasswordChange(id model.UserID) error {
	e.log.Info("user change password", id.Zap())

	if err := e.session.RevokeUser(context.Background(), id); err != nil {
		e.log.Error("failed to revoke user", id.Zap(), zap.Error(err))
		return errgo.Wrap(err, "session.RevokeUser")
	}

	return nil
}

func (e *eventHandler) onMessage(msg kafka.Message) error {
	if len(msg.Value) == 0 {
		// fake event, just ignore
		// https://debezium.io/documentation/reference/stable/connectors/mysql.html#mysql-tombstone-events
		return nil
	}

	var k messageKey
	if err := json.Unmarshal(msg.Key, &k); err != nil {
		return nil
	}

	var v messageValue
	if err := json.Unmarshal(msg.Value, &v); err != nil {
		return nil
	}

	var err error
	switch msg.Topic {
	case "chii.bangumi.chii_subject_fields":
		err = e.OnSubjectField(k.Payload, v.Payload)
	case "chii.bangumi.chii_subjects":
		err = e.OnSubject(k.Payload, v.Payload)
	case "chii.bangumi.chii_members":
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
	Schema  struct {
		Type   string `json:"type"`
		Name   string `json:"name"`
		Fields []struct {
			Type     string `json:"type"`
			Field    string `json:"field"`
			Optional bool   `json:"optional"`
		} `json:"fields"`
		Optional bool `json:"optional"`
	} `json:"schema"`
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
