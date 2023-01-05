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

package handler

import (
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"

	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/revision"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/user"
	"github.com/bangumi/server/web/session"
)

func New(
	a auth.Service,
	r revision.Repo,
	index index.Repo,
	cache cache.RedisCache,
	ctrl ctrl.Ctrl,
	session session.Manager,
	subject subject.Repo,
	search search.Handler,
	u user.Repo,
	episode episode.Repo,
	log *zap.Logger,
) Handler {
	return Handler{
		ctrl:     ctrl,
		episode:  episode,
		cache:    cache,
		log:      log.Named("web.handler"),
		session:  session,
		u:        u,
		subject:  subject,
		a:        a,
		i:        index,
		search:   search,
		r:        r,
		buffPool: buffer.NewPool(),
	}
}

type Handler struct {
	ctrl     ctrl.Ctrl
	episode  episode.Repo
	r        revision.Repo
	cache    cache.RedisCache
	subject  subject.Repo
	a        auth.Service
	session  session.Manager
	u        user.Repo
	i        index.Repo
	search   search.Handler
	buffPool buffer.Pool
	log      *zap.Logger
}
