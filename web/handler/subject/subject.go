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

package subject

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/topic"
	"github.com/bangumi/server/web/handler/common"
)

type Subject struct {
	common.Common
	ctrl   ctrl.Ctrl
	person person.Service
	topic  topic.Repo
	log    *zap.Logger
	cfg    config.AppConfig
}

func New(
	common common.Common,
	p person.Service,
	topic topic.Repo,
	ctrl ctrl.Ctrl,
	log *zap.Logger,
) (Subject, error) {
	return Subject{
		Common: common,
		ctrl:   ctrl,
		person: p,
		topic:  topic,
		log:    log.Named("handler.Subject"),
		cfg:    config.AppConfig{},
	}, nil
}
