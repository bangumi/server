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

package dao

import (
	"time"

	"github.com/bangumi/server/internal/model"
)

func (c *CharacterComment) CreatorID() model.UserID {
	return model.UserID(c.UID)
}

func (c *CharacterComment) IsSubComment() bool {
	return c.Related == 0
}

func (c *CharacterComment) CommentID() model.CommentID {
	return c.ID
}

func (c *CharacterComment) CreateAt() time.Time {
	return time.Unix(int64(c.CreatedTime), 0)
}

func (c *CharacterComment) GetState() uint8 {
	return c.statStub()
}

func (c *CharacterComment) RelatedTo() model.CommentID {
	return c.Related
}

func (c *CharacterComment) GetID() model.CommentID {
	return c.ID
}

func (c *CharacterComment) GetContent() string {
	return c.Content
}

func (c *CharacterComment) GetMentionedID() model.UserID {
	return model.UserID(c.MentionedID)
}

func (c *CharacterComment) statStub() uint8 {
	return 0
}
