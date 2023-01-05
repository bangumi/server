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

package pm

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/internal/user"
	"github.com/bangumi/server/web/handler/common"
)

type PrivateMessage struct {
	ctrl ctrl.Ctrl
	common.Common
	pmRepo pm.Repo
	u      user.Repo
	log    *zap.Logger
}

func New(
	common common.Common,
	pmRepo pm.Repo,
	ctrl ctrl.Ctrl,
	user user.Repo,
	log *zap.Logger,
) (PrivateMessage, error) {
	return PrivateMessage{
		Common: common,
		ctrl:   ctrl,
		u:      user,
		pmRepo: pmRepo,
		log:    log.Named("handler.PrivateMessage"),
	}, nil
}
