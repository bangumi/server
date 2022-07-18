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

package log

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
)

func UserID(id model.UserID) zap.Field {
	return zap.Uint32("user_id", uint32(id))
}

func UserGroup(id model.UserGroupID) zap.Field {
	return zap.Uint8("user_group_id", id)
}

func SubjectID(id model.SubjectID) zap.Field {
	return zap.Uint32("subject_id", uint32(id))
}

func EpisodeID(id model.EpisodeID) zap.Field {
	return zap.Uint32("episode_id", uint32(id))
}

func GroupID(id model.GroupID) zap.Field {
	return zap.Uint16("group_id", uint16(id))
}

func PersonID(id model.PersonID) zap.Field {
	return zap.Uint32("person_id", uint32(id))
}
