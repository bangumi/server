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

type Topic struct {
	CreatedTime time.Time
	UpdatedTime time.Time
	Title       string
	// Comments    []Comment
	ID        TopicID
	CreatorID UserID
	Replies   uint32
	ObjectID  uint32
	State     TopicState
	Status    TopicStatus
}

type TopicState uint8

const (
	TopicStateNone    TopicState = 0 // 正常
	TopicStateClosed  TopicState = 1 // 关闭
	TopicStateReopen  TopicState = 2 // 重开
	TopicStatePin     TopicState = 3 // 置顶
	TopicStateMerge   TopicState = 4 // 合并
	TopicStateSilent  TopicState = 5 // 下沉
	TopicStateDelete  TopicState = 6 // 自行删除
	TopicStatePrivate TopicState = 7 // 管理员删除
)

type TopicStatus uint8

const (
	TopicStatusBan    TopicStatus = 0
	TopicStatusNormal TopicStatus = 1
	TopicStatusReview TopicStatus = 2
)
