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

package cachekey

import (
	"strconv"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/model"
)

const resPrefix = config.RedisKeyPrefix + "repo:"

func Character(id model.CharacterID) string {
	return resPrefix + "character:" + strconv.FormatUint(uint64(id), 10)
}

func Person(id model.PersonID) string {
	return resPrefix + "person:" + strconv.FormatUint(uint64(id), 10)
}

func Subject(id model.SubjectID) string {
	return resPrefix + "subject:" + strconv.FormatUint(uint64(id), 10)
}

func Episode(id model.EpisodeID) string {
	return resPrefix + "episode:" + strconv.FormatUint(uint64(id), 10)
}

func Index(id model.IndexID) string {
	return resPrefix + "index:" + strconv.FormatUint(uint64(id), 10)
}

func User(id model.UserID) string {
	return resPrefix + "user:" + strconv.FormatUint(uint64(id), 10)
}
