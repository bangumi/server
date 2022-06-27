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

import "database/sql/driver"

type EpisodeID uint32
type SubjectID uint32
type PersonID uint32
type CharacterID uint32
type UserID uint32
type GroupID uint16

type TopicID = uint32
type CommentID = uint32
type IndexID = uint32
type RevisionID = uint32
type UserGroupID = uint8
type EpType = int16
type RevisionType = uint8

var _ driver.Valuer = UserID(0)
var _ driver.Valuer = PersonID(0)
var _ driver.Valuer = CharacterID(0)
var _ driver.Valuer = SubjectID(0)
var _ driver.Valuer = EpisodeID(0)
var _ driver.Valuer = GroupID(0)

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
