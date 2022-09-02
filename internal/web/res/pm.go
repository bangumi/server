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

package res

import (
	"time"

	"github.com/bangumi/server/internal/model"
)

type PrivateMessage struct {
	CreatedTime    time.Time       `json:"created_time"`
	RelatedMessage *PrivateMessage `json:"related_message"`
	Sender         *User           `json:"sender"`
	Receiver       *User           `json:"receiver"`
	Title          string          `json:"title"`
	Content        string          `json:"content"`
	ID             uint32          `json:"id"`
	New            bool            `json:"new"`
}

type PrivateMessageTypeCounts struct {
	Unread int64 `json:"unread"`
	Inbox  int64 `json:"inbox"`
	Outbox int64 `json:"outbox"`
}

func ConvertModelPrivateMessage(
	item model.PrivateMessage,
	users map[model.UserID]model.User,
) PrivateMessage {
	msg := PrivateMessage{
		CreatedTime: item.CreatedTime,
		Title:       item.Title,
		Content:     item.Content,
		ID:          uint32(item.ID),
		New:         item.New,
	}
	if users != nil {
		if u, ok := users[item.SenderID]; ok {
			user := ConvertModelUser(u)
			msg.Sender = &user
		}
		if u, ok := users[item.ReceiverID]; ok {
			user := ConvertModelUser(u)
			msg.Receiver = &user
		}
	}
	return msg
}

func ConvertModelPrivateMessageListItem(
	item model.PrivateMessageListItem,
	users map[model.UserID]model.User,
) PrivateMessage {
	relatedMsg := ConvertModelPrivateMessage(item.Main, nil)
	msg := ConvertModelPrivateMessage(item.Self, users)
	msg.RelatedMessage = &relatedMsg
	return msg
}
