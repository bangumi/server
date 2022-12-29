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
	"go.uber.org/fx"

	"github.com/bangumi/server/web/handler/character"
	"github.com/bangumi/server/web/handler/common"
	"github.com/bangumi/server/web/handler/index"
	"github.com/bangumi/server/web/handler/notification"
	"github.com/bangumi/server/web/handler/person"
	"github.com/bangumi/server/web/handler/pm"
	"github.com/bangumi/server/web/handler/subject"
	"github.com/bangumi/server/web/handler/user"
)

var Module = fx.Module("handler",
	fx.Provide(
		New,
		common.New,
		user.New,
		person.New,
		subject.New,
		character.New,
		index.New,
		pm.New,
		notification.New,
	),
)
