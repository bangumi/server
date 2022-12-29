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

package notification

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/internal/notification"
	"github.com/bangumi/server/web/handler/common"
)

type Notification struct {
	ctrl ctrl.Ctrl
	common.Common
	notificationRepo notification.Repo
	log              *zap.Logger
}

func New(
	common common.Common,
	notificationRepo notification.Repo,
	ctrl ctrl.Ctrl,
	log *zap.Logger,
) (Notification, error) {
	return Notification{
		Common:           common,
		ctrl:             ctrl,
		notificationRepo: notificationRepo,
		log:              log.Named("handler.Notification"),
	}, nil
}
