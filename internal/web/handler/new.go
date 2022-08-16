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

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/ctrl"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/oauth"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/web/captcha"
	"github.com/bangumi/server/internal/web/frontend"
	"github.com/bangumi/server/internal/web/handler/common"
	"github.com/bangumi/server/internal/web/rate"
	"github.com/bangumi/server/internal/web/session"
)

func New(
	common common.Common,
	a domain.AuthService,
	r domain.RevisionRepo,
	g domain.GroupRepo,
	index domain.IndexRepo,
	cache cache.Cache,
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
	g         domain.GroupRepo
	oauth     oauth.Manager
	r         domain.RevisionRepo
	cache     cache.Cache
	a         domain.AuthService
	session   session.Manager
	captcha   captcha.Manager
	rateLimit rate.Manager
	i         domain.IndexRepo
	search    search.Handler
	template  frontend.TemplateEngine
	buffPool  buffer.Pool
	log       *zap.Logger
}
