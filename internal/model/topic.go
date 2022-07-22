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
	CreatedAt time.Time
	UpdatedAt time.Time
	Title     string
	ID        TopicID
	CreatorID UserID
	Replies   uint32
	ObjectID  uint32
	State     CommentState
	Status    TopicStatus
}

type CommentState uint8

const (
	CommentStateNone CommentState = 0 // 正常
	// CommentStateAdminCloseTopic 管理员关闭主题 https://bgm.tv/subject/topic/12629#post_108127
	CommentStateAdminCloseTopic CommentState = 1 // 关闭
	CommentStateAdminReopen     CommentState = 2 // 重开
	CommentStateAdminPin        CommentState = 3 // 置顶
	CommentStateAdminMerge      CommentState = 4 // 合并
	// CommentStateAdminSilentTopic 管理员下沉 https://bgm.tv/subject/topic/18784#post_160402
	CommentStateAdminSilentTopic CommentState = 5 // 下沉
	CommentStateUserDelete       CommentState = 6 // 自行删除
	CommentStateAdminDelete      CommentState = 7 // 管理员删除
)

type TopicStatus uint8

const (
	TopicStatusBan    TopicStatus = 0
	TopicStatusNormal TopicStatus = 1
	TopicStatusReview TopicStatus = 2
)
