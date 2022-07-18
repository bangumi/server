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

package person

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/app"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/web/handler/common"
)

type Person struct {
	common.Common
	app   app.App
	topic domain.TopicRepo
	log   *zap.Logger
}

func New(
	common common.Common,
	topic domain.TopicRepo,
	app app.App,
	log *zap.Logger,
) (Person, error) {
	return Person{
		Common: common,
		app:    app,
		topic:  topic,
		log:    log.Named("handler.Person"),
	}, nil
}
