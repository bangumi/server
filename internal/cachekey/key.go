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

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/model"
)

// Put version in cache key to avoid model changes.
const globalPrefix = "chii:" + config.Version + ":res:"

func Character(id model.CharacterID) string {
	return globalPrefix + "character:" + strconv.FormatUint(uint64(id), 10)
}

func Person(id model.PersonID) string {
	return globalPrefix + "person:" + strconv.FormatUint(uint64(id), 10)
}

func Subject(id model.SubjectID) string {
	return globalPrefix + "subject:" + strconv.FormatUint(uint64(id), 10)
}

func Episode(id model.EpisodeID) string {
	return globalPrefix + "episode:" + strconv.FormatUint(uint64(id), 10)
}

func Index(id model.IndexID) string {
	return globalPrefix + "index:" + strconv.FormatUint(uint64(id), 10)
}

func Auth(token string) string {
	return "chii:" + config.Version + ":auth:access-token:" + token
}

func User(id model.UserID) string {
	return globalPrefix + "user:" + strconv.FormatUint(uint64(id), 10)
}
