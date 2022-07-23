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

type PrivateTopic struct {
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	Title      string        `json:"title"`
	Creator    User          `json:"creator"`
	ID         model.TopicID `json:"id"`
	ReplyCount uint32        `json:"reply_count"`
}

type PrivateTopicDetail struct {
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	Title      string           `json:"title"`
	Creator    User             `json:"creator"`
	Text       string           `json:"text"`
	Comments   []PrivateComment `json:"comments"`
	ID         model.TopicID    `json:"id"`
	IsFriend   bool             `json:"is_friend"`
	State      CommentState     `json:"state"`
	ReplyCount uint32           `json:"reply_count"`
}

type PrivateComments struct {
	Comments []PrivateComment `json:"comments"`
}

type PrivateComment struct {
	CreatedAt time.Time           `json:"created_at"`
	Text      string              `json:"text"`
	Creator   User                `json:"creator"`
	Replies   []PrivateSubComment `json:"replies"`
	ID        model.CommentID     `json:"id"`
	IsFriend  bool                `json:"is_friend"`
	State     CommentState        `json:"state"`
}

type PrivateSubComment struct {
	CreatedAt time.Time       `json:"created_at"`
	Text      string          `json:"text"`
	Creator   User            `json:"creator"`
	IsFriend  bool            `json:"is_friend"`
	State     CommentState    `json:"state"`
	ID        model.CommentID `json:"id"`
}

type CommentState uint8

const (
	CommentNormal           = CommentState(model.CommentStateNone)
	CommentAdminCloseTopic  = CommentState(model.CommentStateAdminCloseTopic)
	CommentAdminSilentTopic = CommentState(model.CommentStateAdminSilentTopic)
	CommentDeletedByUser    = CommentState(model.CommentStateUserDelete)
	CommentDeletedByAdmin   = CommentState(model.CommentStateAdminDelete)
)

func ToCommentState(i model.CommentState) CommentState {
	switch i {
	case model.CommentStateNone:
		return CommentNormal
	case model.CommentStateAdminDelete:
		return CommentDeletedByAdmin
	case model.CommentStateAdminCloseTopic:
		return CommentAdminCloseTopic
	case model.CommentStateUserDelete:
		return CommentDeletedByUser
	case model.CommentStateAdminSilentTopic:
		return CommentAdminSilentTopic
	default:
		return CommentNormal
	}
}
