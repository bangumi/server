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

package character

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/collections"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/subject"
)

type Character struct {
	ctrl      ctrl.Ctrl
	person    person.Service
	character character.Repo
	subject   subject.Repo
	collect   collections.Repo
	log       *zap.Logger
	cfg       config.AppConfig
}

func New(
	person person.Service,
	ctrl ctrl.Ctrl,
	character character.Repo,
	subject subject.Repo,
	collect collections.Repo,
	log *zap.Logger,
) (Character, error) {
	return Character{
		ctrl:      ctrl,
		character: character,
		subject:   subject,
		person:    person,
		collect:   collect,
		log:       log.Named("handler.Character"),
		cfg:       config.AppConfig{},
	}, nil
}
