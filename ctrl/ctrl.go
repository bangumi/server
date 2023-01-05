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

	"github.com/bangumi/server/dal"
	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/dam"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/timeline"
	"github.com/bangumi/server/internal/user"
)

func New(
	episode episode.Repo,
	cache cache.RedisCache,
	subject subject.Repo,
	person person.Repo,
	collection collection.Repo,
	timeline timeline.Repo,
	user user.Repo,
	tx dal.Transaction,
	dam dam.Dam,
	privateMessage pm.Repo,
	log *zap.Logger,
) Ctrl {
	return Ctrl{
		log:   log.Named("controller"),
		cache: cache,

		tx:  tx,
		dam: dam,

		user:           user,
		person:         person,
		episode:        episode,
		subject:        subject,
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

	user           user.Repo
	person         person.Repo
	episode        episode.Repo
	subject        subject.Repo
	collection     collection.Repo
	timeline       timeline.Repo
	privateMessage pm.Repo
}
