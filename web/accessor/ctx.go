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

package accessor

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/web/internal/ctxkey"
)

func NewFromCtx(c echo.Context) *Accessor {
	a := get()
	a.fillBasicInfo(c)
	return a
}

func GetFromCtx(c echo.Context) *Accessor {
	raw := c.Get(ctxkey.User)
	if raw == nil {
		return NewFromCtx(c)
	}

	u, ok := raw.(*Accessor)
	if !ok {
		logger.Error("failed to get http accessor, expecting accessor got another type instead", zap.Any("raw", raw))
		panic("can't convert type")
	}

	return u
}
