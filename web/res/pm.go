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

	"github.com/samber/lo"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/internal/user"
)

type PrivateMessage struct {
	CreatedAt      time.Time       `json:"created_at"`
	RelatedMessage *PrivateMessage `json:"related_message,omitempty"`
	Sender         *User           `json:"sender,omitempty"`
	Receiver       *User           `json:"receiver,omitempty"`
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

func ConvertModelPrivateMessage(item pm.PrivateMessage, users map[model.UserID]user.User) PrivateMessage {
	msg := PrivateMessage{
		CreatedAt: item.CreatedTime,
		Title:     item.Title,
		Content:   item.Content,
		ID:        item.ID,
		New:       item.New,
	}
	if users != nil {
		if u, ok := users[item.SenderID]; ok {
			msg.Sender = lo.ToPtr(ConvertModelUser(u))
		}
		if u, ok := users[item.ReceiverID]; ok {
			msg.Receiver = lo.ToPtr(ConvertModelUser(u))
		}
	}
	return msg
}

func ConvertModelPrivateMessageListItem(
	item pm.PrivateMessageListItem,
	users map[model.UserID]user.User,
) PrivateMessage {
	relatedMsg := ConvertModelPrivateMessage(item.Main, nil)
	msg := ConvertModelPrivateMessage(item.Self, users)
	msg.RelatedMessage = &relatedMsg
	return msg
}
