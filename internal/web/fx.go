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

package web

import (
	"go.uber.org/fx"

	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/web/captcha"
	"github.com/bangumi/server/internal/web/frontend"
	"github.com/bangumi/server/internal/web/handler"
	"github.com/bangumi/server/internal/web/rate"
	"github.com/bangumi/server/internal/web/session"
)

var Module = fx.Module("web",
	handler.Module,
	fx.Provide(
		New,
		session.NewMysqlRepo,
		rate.New,
		captcha.New,
		session.New,
		frontend.NewTemplateEngine,
		func(c *search.Client) search.Handler { return c },
	),
	fx.Invoke(AddRouters),
)
