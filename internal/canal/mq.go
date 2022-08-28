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

//nolint:depguard
package canal

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"

	"github.com/bangumi/server/internal/pkg/logger"
)

func Main() error {
	e, err := getEventHandler()
	if err != nil {
		return err
	}

	shutdown := make(chan int)
	sigChan := make(chan os.Signal, 1)
	// register for interrupt (Ctrl+C) and SIGTERM (docker)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("receive signal, shutdown")
		e.Close()
		shutdown <- 1
	}()

	select {
	case err := <-e.start(context.Background()):
		return err
	case <-shutdown:
		return nil
	}
}

const groupID = "my-group"

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
	Before map[string]json.RawMessage `json:"before"`
	After  map[string]json.RawMessage `json:"after"`
	Source source                     `json:"source"`
	Op     string                     `json:"op"`
}

type source struct {
	Table string `json:"table"`
}
