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

type Topic struct {
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	Title      string        `json:"title"`
	Creator    User          `json:"creator"`
	ID         model.TopicID `json:"id"`
	ReplyCount uint32        `json:"reply_count"`
}

type TopicDetail struct {
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	Title      string          `json:"title"`
	Creator    User            `json:"creator"`
	Text       string          `json:"text"`
	Comments   PagedG[Comment] `json:"comments,omitempty"`
	ID         model.TopicID   `json:"id"`
	ReplyCount uint32          `json:"reply_count"`
}

type Comment struct {
	CreatedAt time.Time       `json:"created_at"`
	Replies   []SubComment    `json:"replies"`
	Text      string          `json:"text"`
	Creator   User            `json:"creator"`
	ID        model.CommentID `json:"id"`
}

type SubComment struct {
	CreatedAt time.Time       `json:"created_at"`
	Text      string          `json:"text"`
	Creator   User            `json:"creator"`
	ID        model.CommentID `json:"id"`
}
