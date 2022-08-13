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
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type K struct {
}

func Main2() error {
	go produce(context.Background())
	consume(context.Background())
	return nil
}

// the topic and broker address are initialized as constants
const (
	topic          = "message-log"
	broker1Address = "192.168.1.3:29092"
)

func produce(ctx context.Context) {
	// initialize a counter
	i := 0

	w := &kafka.Writer{
		Addr:                   kafka.TCP("192.168.1.3:29092"),
		Topic:                  topic,
		AllowAutoTopicCreation: true,
	}

	for {
		// each kafka message has a key and value. The key is used
		// to decide which partition (and consequently, which broker)
		// the message gets published on
		err := w.WriteMessages(ctx, kafka.Message{
			Key: []byte(strconv.Itoa(i)),
			// create an arbitrary message payload for the value
			Value: []byte("this is message " + strconv.Itoa(i)),
		})
		if err != nil {
			panic("could not write message " + err.Error())
		}

		// log a confirmation once the message is written
		fmt.Println("writes:", i)
		i++
		// sleep for a second
		time.Sleep(time.Second)
	}
}

func consume(ctx context.Context) {
	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker1Address},
		GroupID:     "my-group",
		GroupTopics: nil,
		Topic:       topic,
	})

	for {
		// the `ReadMessage` method blocks until we receive the next event
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		// after receiving the message, log its value
		fmt.Println("received: ", string(msg.Value))
	}
}
