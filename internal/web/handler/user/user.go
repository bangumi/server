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

package user

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/app"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/web/handler/common"
)

type User struct {
	common.Common
	app     app.App
	person  domain.PersonService
	topic   domain.TopicRepo
	collect domain.CollectionRepo
	log     *zap.Logger
	user    domain.UserRepo
	cfg     config.AppConfig
}

func New(
	common common.Common,
	p domain.PersonService,
	user domain.UserRepo,
	topic domain.TopicRepo,
	app app.App,
	collect domain.CollectionRepo,
	log *zap.Logger,
) (User, error) {
	return User{
		Common:  common,
		app:     app,
		collect: collect,
		user:    user,
		person:  p,
		topic:   topic,
		log:     log.Named("handler.User"),
		cfg:     config.AppConfig{},
	}, nil
}
