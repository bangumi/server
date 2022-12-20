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

import "time"

type NotificationField struct {
	Title       string
	ID          NotificationFieldID
	RelatedID   uint32 // 关联的主体事项的id，条目讨论，日志的
	RelatedType uint8  // 关联的主体事项的分类
}

type Notification struct {
	CreatedTime time.Time
	ID          NotificationID
	ReceiverID  UserID
	SenderID    UserID
	FieldID     NotificationFieldID
	RelatedID   uint32 // 触发通知的实际事项id，如回复的
	Type        NotificationType
	Status      NotificationStatus
}
