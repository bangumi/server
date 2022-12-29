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

	"github.com/bangumi/server/internal/ctrl"
	"github.com/bangumi/server/internal/topic"
	"github.com/bangumi/server/internal/web/handler/common"
)

type Person struct {
	common.Common
	ctrl  ctrl.Ctrl
	topic topic.Repo
	log   *zap.Logger
}

func New(
	common common.Common,
	topic topic.Repo,
	ctrl ctrl.Ctrl,
	log *zap.Logger,
) (Person, error) {
	return Person{
		Common: common,
		ctrl:   ctrl,
		topic:  topic,
		log:    log.Named("handler.Person"),
	}, nil
}
