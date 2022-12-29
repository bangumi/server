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
)

func (t *GroupTopic) GetTitle() string {
	return t.Title
}

func (t *SubjectTopic) GetTitle() string {
	return t.Title
}

func (t *SubjectTopic) GetCreateTime() time.Time {
	return time.Unix(int64(t.CreatedTime), 0)
}

func (t *SubjectTopic) GetUpdateTime() time.Time {
	return time.Unix(int64(t.UpdatedTime), 0)
}

func (t *SubjectTopic) GetID() uint32 {
	return t.ID
}

func (t *SubjectTopic) GetCreatorID() uint32 {
	return t.UID
}

func (t *SubjectTopic) GetState() uint8 {
	return t.State
}

func (t *GroupTopic) GetCreateTime() time.Time {
	return time.Unix(int64(t.CreatedTime), 0)
}

func (t *GroupTopic) GetUpdateTime() time.Time {
	return time.Unix(int64(t.UpdatedTime), 0)
}

func (t *GroupTopic) GetID() uint32 {
	return t.ID
}

func (t *GroupTopic) GetCreatorID() uint32 {
	return t.UID
}

func (t *GroupTopic) GetState() uint8 {
	return t.State
}

func (t *GroupTopic) GetReplies() uint32 {
	return t.Replies
}

func (t *SubjectTopic) GetReplies() uint32 {
	return t.Replies
}

func (t *SubjectTopic) GetParentID() uint32 {
	return t.SubjectID
}

func (t *GroupTopic) GetParentID() uint32 {
	return t.GroupID
}

func (t *SubjectTopic) GetDisplay() uint8 {
	return t.Display
}

func (t *GroupTopic) GetDisplay() uint8 {
	return t.Display
}
