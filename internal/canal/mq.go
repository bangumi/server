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

	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
)

type streamReader struct {
	handler func(key json.RawMessage, payload payload)
	Topic   string
}

func Main() error {
	var eg errgroup.Group
	closers, err := startReaders(&eg)
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
		for _, closer := range closers {
			closer.Close()
		}
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

func startReaders(eg *errgroup.Group) ([]io.Closer, error) {
	cfg, err := config.NewAppConfig()
	if err != nil {
		return nil, errgo.Wrap(err, "config.NewAppConfig")
	}

	e, err := getEventHandler()
	if err != nil {
		return nil, err
	}

	var closers = make([]io.Closer, 0)

	for _, readerCfg := range []streamReader{
		{Topic: "chii.bangumi.chii_subjects", handler: OnSubjectChange},
		{Topic: "chii.bangumi.chii_members", handler: e.OnUserChange},
	} {
		readerCfg := readerCfg
		eg.Go(func() error {
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers:     []string{cfg.KafkaBroker},
				GroupID:     "my-group",
				GroupTopics: nil,
				Topic:       readerCfg.Topic,
			})

			closers = append(closers, reader)

			for {
				msg, err := reader.ReadMessage(context.Background())
				if err != nil {
					return errgo.Wrap(err, "reader.ReadMessage")
				}

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

				readerCfg.handler(k.Payload, v.Payload)
			}
		})
	}

	return closers, nil
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
