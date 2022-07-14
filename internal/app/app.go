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

package app

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/oauth"
)

var Module = fx.Module("app", fx.Provide(New))

func New(
	cfg config.AppConfig,
	subject domain.SubjectService,
	characterService domain.CharacterService,
	personService domain.PersonService,
	auth domain.AuthService,
	episode domain.EpisodeRepo,
	collect domain.CollectionRepo,
	revision domain.RevisionRepo,
	topic domain.TopicRepo,
	comment domain.CommentRepo,
	group domain.GroupRepo,
	index domain.IndexRepo,
	user domain.UserRepo,
	cache cache.Generic,
	log *zap.Logger,
	oauth oauth.Manager,
) (App, error) {
	log = log.Named("App")

	return App{
		cfg:       cfg,
		cache:     cache,
		log:       log,
		person:    personService,
		subject:   subject,
		auth:      auth,
		user:      user,
		episode:   episode,
		character: characterService,
		collect:   collect,
		index:     index,
		revision:  revision,
		topic:     topic,
		comment:   comment,
		oauth:     oauth,
		group:     group,
		buffPool:  buffer.NewPool(),
	}, nil
}

type App struct {
	subject   domain.SubjectService
	person    domain.PersonService
	auth      domain.AuthService
	collect   domain.CollectionRepo
	episode   domain.EpisodeRepo
	oauth     oauth.Manager
	character domain.CharacterService
	user      domain.UserRepo
	revision  domain.RevisionRepo
	index     domain.IndexRepo
	comment   domain.CommentRepo
	topic     domain.TopicRepo
	group     domain.GroupRepo
	cache     cache.Generic
	log       *zap.Logger
	buffPool  buffer.Pool
	cfg       config.AppConfig
}
