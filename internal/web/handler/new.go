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
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/ctrl"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/oauth"
	"github.com/bangumi/server/internal/web/captcha"
	"github.com/bangumi/server/internal/web/frontend"
	"github.com/bangumi/server/internal/web/handler/character"
	"github.com/bangumi/server/internal/web/handler/common"
	"github.com/bangumi/server/internal/web/handler/person"
	"github.com/bangumi/server/internal/web/handler/subject"
	"github.com/bangumi/server/internal/web/handler/user"
	"github.com/bangumi/server/internal/web/rate"
	"github.com/bangumi/server/internal/web/session"
)

func New(
	common common.Common,
	cfg config.AppConfig,
	a domain.AuthService,
	r domain.RevisionRepo,
	topic domain.TopicRepo,
	g domain.GroupRepo,
	index domain.IndexRepo,
	user domain.UserRepo,
	cache cache.Cache,
	ctrl ctrl.Ctrl,
	captcha captcha.Manager,
	session session.Manager,
	rateLimit rate.Manager,
	userHandler user.User,
	personHandler person.Person,
	log *zap.Logger,
	subject subject.Subject,
	engine frontend.TemplateEngine,
	character character.Character,
	oauth oauth.Manager,
) (Handler, error) {
	return Handler{
		Subject:   subject,
		Common:    common,
		ctrl:      ctrl,
		User:      userHandler,
		Character: character,
		Person:    personHandler,
		cfg:       cfg,
		cache:     cache,
		log:       log.Named("web.handler"),
		rateLimit: rateLimit,
		session:   session,
		a:         a,
		u:         user,
		i:         index,
		r:         r,
		topic:     topic,
		captcha:   captcha,
		g:         g,
		oauth:     oauth,
		template:  engine,
		buffPool:  buffer.NewPool(),
	}, nil
}

type Handler struct {
	common.Common
	Subject   subject.Subject
	Character character.Character
	Person    person.Person
	ctrl      ctrl.Ctrl
	User      user.User
	a         domain.AuthService
	session   session.Manager
	captcha   captcha.Manager
	u         domain.UserRepo
	rateLimit rate.Manager
	i         domain.IndexRepo
	g         domain.GroupRepo
	cache     cache.Cache
	r         domain.RevisionRepo
	oauth     oauth.Manager
	topic     domain.TopicRepo
	template  frontend.TemplateEngine
	buffPool  buffer.Pool
	log       *zap.Logger
	cfg       config.AppConfig
}
