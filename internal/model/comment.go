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

type Comment struct {
	CreatedAt   time.Time
	Content     string
	SubComments []SubComment
	CreatorID   UserID
	ID          CommentID
	State       uint8
}

type SubComment struct {
	CreatedAt time.Time
	Content   string
	CreatorID UserID
	Related   CommentID
	State     uint8
	ID        CommentID
	ObjectID  TopicID
}
