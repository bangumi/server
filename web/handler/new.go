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
	"github.com/bangumi/server/internal/group"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/oauth"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/revision"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/web/captcha"
	"github.com/bangumi/server/web/frontend"
	"github.com/bangumi/server/web/handler/common"
	"github.com/bangumi/server/web/rate"
	"github.com/bangumi/server/web/session"
)

func New(
	common common.Common,
	a auth.Service,
	r revision.Repo,
	g group.Repo,
	index index.Repo,
	cache cache.RedisCache,
	ctrl ctrl.Ctrl,
	captcha captcha.Manager,
	session session.Manager,
	rateLimit rate.Manager,
	search search.Handler,
	log *zap.Logger,
	engine frontend.TemplateEngine,
	oauth oauth.Manager,
) Handler {
	return Handler{
		Common:    common,
		ctrl:      ctrl,
		cache:     cache,
		log:       log.Named("web.handler"),
		rateLimit: rateLimit,
		session:   session,
		a:         a,
		i:         index,
		search:    search,
		r:         r,
		captcha:   captcha,
		g:         g,
		oauth:     oauth,
		template:  engine,
		buffPool:  buffer.NewPool(),
	}
}

type Handler struct {
	ctrl ctrl.Ctrl
	common.Common
	g         group.Repo
	oauth     oauth.Manager
	r         revision.Repo
	cache     cache.RedisCache
	a         auth.Service
	session   session.Manager
	captcha   captcha.Manager
	rateLimit rate.Manager
	i         index.Repo
	search    search.Handler
	template  frontend.TemplateEngine
	buffPool  buffer.Pool
	log       *zap.Logger
}
