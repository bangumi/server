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
	"sync"

	"github.com/segmentio/kafka-go"
)

type streamReader struct {
	Topic   string
	handler func(key json.RawMessage, payload payload)
}

func Main() error {
	e, err := getEventHandler()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, readerCfg := range []streamReader{
		{Topic: "chii.bangumi.chii_subjects", handler: OnSubjectChange},
		{Topic: "chii.bangumi.chii_members", handler: e.OnUserChange},
	} {
		readerCfg := readerCfg
		wg.Add(1)
		go func() {
			defer wg.Done()

			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers:     []string{broker1Address},
				GroupID:     "my-group",
				GroupTopics: nil,
				Topic:       readerCfg.Topic,
				MaxBytes:    1024 * 10,
			})

			for {
				msg, err := reader.ReadMessage(context.Background())
				if err != nil {
					break
				}

				if len(msg.Value) == 0 {
					// fake event, just ignore
					// https://debezium.io/documentation/reference/stable/connectors/mysql.html#mysql-tombstone-events
					continue
				}

				var k messageKey
				if err := json.Unmarshal(msg.Key, &k); err != nil {
					fmt.Println(err)
					fmt.Println(string(msg.Key), string(msg.Value))
					continue
				}

				var v messageValue
				if err := json.Unmarshal(msg.Value, &v); err != nil {
					fmt.Println(err)
					fmt.Println(string(msg.Key), string(msg.Value))
					continue
				}

				readerCfg.handler(k.Payload, v.Payload)
			}
		}()

	}

	wg.Wait()
	return nil
}

const (
	opCreate  = "c"
	opReplace = "r"
	opDelete  = "d"
	opUpdate  = "u"
)

// https://debezium.io/documentation/reference/connectors/mysql.html
// Table 9. Overview of change event basic content

type messageKey struct {
	Schema struct {
		Type   string `json:"type"`
		Fields []struct {
			Type     string `json:"type"`
			Optional bool   `json:"optional"`
			Field    string `json:"field"`
		} `json:"fields"`
		Optional bool   `json:"optional"`
		Name     string `json:"name"`
	} `json:"schema"`
	Payload json.RawMessage `json:"payload"`
}

type messageValue struct {
	Payload payload `json:"payload"`
}

type payload struct {
	Before map[string]json.RawMessage `json:"before"`
	After  map[string]json.RawMessage `json:"after"`
	Source source                     `json:"source"`
	Op     string                     `json:"op"`
	TsMs   float64                    `json:"ts_ms"`
}

type source struct {
	Table string `json:"table"`
}
