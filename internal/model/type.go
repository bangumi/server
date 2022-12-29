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
	"database/sql/driver"

	"go.uber.org/zap"
)

type EpisodeID uint32
type SubjectID uint32
type PersonID uint32
type CharacterID uint32
type UserID uint32
type GroupID uint16
type TimeLineID uint32

type TopicID uint32
type CommentID uint32

type IndexID = uint32
type RevisionID = uint32
type RevisionType = uint8
type TimeLineCat = uint16

type PrivateMessageID uint32
type NotificationID uint32
type NotificationFieldID uint32

var _ driver.Valuer = UserID(0)
var _ driver.Valuer = PersonID(0)
var _ driver.Valuer = CharacterID(0)
var _ driver.Valuer = SubjectID(0)
var _ driver.Valuer = EpisodeID(0)
var _ driver.Valuer = GroupID(0)
var _ driver.Valuer = TimeLineID(0)

func (v UserID) Value() (driver.Value, error) {
	return int64(v), nil
}

func (v PersonID) Value() (driver.Value, error) {
	return int64(v), nil
}

func (v CharacterID) Value() (driver.Value, error) {
	return int64(v), nil
}

func (v SubjectID) Value() (driver.Value, error) {
	return int64(v), nil
}

func (v EpisodeID) Value() (driver.Value, error) {
	return int64(v), nil
}

func (v GroupID) Value() (driver.Value, error) {
	return int64(v), nil
}

func (v PrivateMessageID) Value() (driver.Value, error) {
	return int64(v), nil
}

func (v TimeLineID) Value() (driver.Value, error) {
	return int64(v), nil
}

func (v GroupID) Zap() zap.Field {
	return zap.Uint32("group_id", uint32(v))
}

func (v UserID) Zap() zap.Field {
	return zap.Uint32("user_id", uint32(v))
}

func (v SubjectID) Zap() zap.Field {
	return zap.Uint32("subject_id", uint32(v))
}

func (v EpisodeID) Zap() zap.Field {
	return zap.Uint32("episode_id", uint32(v))
}

func (v TopicID) Zap() zap.Field {
	return zap.Uint32("topic_id", uint32(v))
}
