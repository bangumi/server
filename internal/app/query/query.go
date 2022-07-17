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

package query

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/domain"
)

func New(
	episode domain.EpisodeRepo,
	cache cache.Cache,
	subject domain.SubjectRepo,
	log *zap.Logger,
) Query {
	return Query{
		log:     log.Named("app.query"),
		cache:   cache,
		subject: subject,
		episode: episode,
	}
}

type Query struct {
	cache   cache.Cache
	episode domain.EpisodeRepo
	subject domain.SubjectRepo
	log     *zap.Logger
}
