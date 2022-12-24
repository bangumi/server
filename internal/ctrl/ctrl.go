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

package ctrl

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/dam"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/subject"
)

func New(
	episode domain.EpisodeRepo,
	cache cache.RedisCache,
	subject subject.Repo,
	person domain.PersonRepo,
	character domain.CharacterRepo,
	collection domain.CollectionRepo,
	index domain.IndexRepo,
	timeline domain.TimeLineRepo,
	user domain.UserRepo,
	topic domain.TopicRepo,
	tx dal.Transaction,
	dam dam.Dam,
	privateMessage domain.PrivateMessageRepo,
	log *zap.Logger,
) Ctrl {
	return Ctrl{
		log:   log.Named("controller"),
		cache: cache,

		tx:  tx,
		dam: dam,

		user:           user,
		topic:          topic,
		person:         person,
		episode:        episode,
		subject:        subject,
		character:      character,
		index:          index,
		collection:     collection,
		timeline:       timeline,
		privateMessage: privateMessage,
	}
}

type Ctrl struct {
	log   *zap.Logger
	cache cache.RedisCache

	tx  dal.Transaction
	dam dam.Dam

	user           domain.UserRepo
	topic          domain.TopicRepo
	person         domain.PersonRepo
	episode        domain.EpisodeRepo
	subject        subject.Repo
	character      domain.CharacterRepo
	collection     domain.CollectionRepo
	index          domain.IndexRepo
	timeline       domain.TimeLineRepo
	privateMessage domain.PrivateMessageRepo
}
