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
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"golang.org/x/sync/errgroup"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
)

func Main() error {
	var eg errgroup.Group
	closer, err := startReaders(&eg)
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
		closer.Close()
		shutdown <- 1
	}()

	errChan := make(chan error)
	go func() {
		errChan <- eg.Wait()
	}()

	select {
	case err := <-errChan:
		return err
	case <-shutdown:
		return nil
	}
}

const groupID = "my-group"

func startReaders(eg *errgroup.Group) (io.Closer, error) {
	appConfig, err := config.NewAppConfig()
	if err != nil {
		return nil, errgo.Wrap(err, "config.NewAppConfig")
	}

	e, err := getEventHandler()
	if err != nil {
		return nil, err
	}

	cfg := sarama.NewConfig()
	client, err := sarama.NewConsumerGroup([]string{appConfig.KafkaBroker}, groupID, cfg)
	if err != nil {
		panic(err)
	}

	topics := []string{"chii.bangumi.chii_subject_fields", "chii.bangumi.chii_subjects", "chii.bangumi.chii_members"}

	eg.Go(func() error { return client.Consume(context.Background(), topics, e) }) //nolint:wrapcheck

	return client, nil
}

func (e *eventHandler) Setup(groupSession sarama.ConsumerGroupSession) error {
	e.log.Info("eventHandler sarama setup")
	return nil
}

func (e *eventHandler) Cleanup(groupSession sarama.ConsumerGroupSession) error {
	e.log.Info("eventHandler sarama cleanup")
	return nil
}

func (e *eventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	e.log.Info("eventHandler sarama ConsumeClaim")
	for {
		select {
		case msg := <-claim.Messages():
			if len(msg.Value) == 0 {
				// fake event, just ignore
				// https://debezium.io/documentation/reference/stable/connectors/mysql.html#mysql-tombstone-events
				continue
			}

			var k messageKey
			if err := json.Unmarshal(msg.Key, &k); err != nil {
				continue
			}

			var v messageValue
			if err := json.Unmarshal(msg.Value, &v); err != nil {
				continue
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

			if err != nil {
				continue
			}

			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
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
