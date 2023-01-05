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

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/user"
)

type User struct {
	ctrl    ctrl.Ctrl
	episode episode.Repo
	person  person.Service
	collect collection.Repo
	subject subject.CachedRepo
	log     *zap.Logger
	user    user.Repo
	cfg     config.AppConfig
}

func New(
	p person.Service,
	user user.Repo,
	ctrl ctrl.Ctrl,
	subject subject.Repo,
	collect collection.Repo,
	episode episode.Repo,
	log *zap.Logger,
) (User, error) {
	return User{
		ctrl:    ctrl,
		episode: episode,
		collect: collect,
		subject: subject,
		user:    user,
		person:  p,
		log:     log.Named("handler.User"),
		cfg:     config.AppConfig{},
	}, nil
}
