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
package model

import (
	"time"
)

type PrivateMessage struct {
	CreatedTime       time.Time
	Title             string
	Content           string
	Folder            PrivateMessageFolderType
	SenderID          UserID
	ReceiverID        UserID
	ID                PrivateMessageID
	MainMessageID     PrivateMessageID // 如果当前是首条私信，则为当前私信的id，否则为0
	RelatedMessageID  PrivateMessageID // 首条私信的id
	New               bool
	DeletedBySender   bool
	DeletedByReceiver bool
}

type PrivateMessageListItem struct {
	Main PrivateMessage
	Self PrivateMessage
}

type PrivateMessageTypeCounts struct {
	Unread int64
	Inbox  int64
	Outbox int64
}
