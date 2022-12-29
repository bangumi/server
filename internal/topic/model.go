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

package topic

import (
	"go.uber.org/zap"
)

type Type uint32

func (t Type) Zap() zap.Field {
	return zap.Uint32("topic_type", uint32(t))
}

const (
	TypeUnknown Type = iota
	TypeSubject
	TypeGroup
)

type CommentType uint32

const (
	CommentTypeUnknown CommentType = iota
	CommentTypeSubjectTopic
	CommentTypeGroupTopic
	CommentIndex
	CommentCharacter
	CommentPerson
	CommentEpisode
)
