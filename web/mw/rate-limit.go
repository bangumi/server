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

package mw

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/rueidis"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/web/res"
)

//go:embed ban.lua
var rateLimitLua string

func RateLimit(cfg config.AppConfig, r rueidis.Client) echo.MiddlewareFunc {
	script := rueidis.NewLuaScript(rateLimitLua)

	args := []string{
		fmt.Sprintf("%d", cfg.RateLimit.LimitLongTime/time.Second),
		fmt.Sprintf("%d", cfg.RateLimit.LimitWindow/time.Second),
		fmt.Sprintf("%d", cfg.RateLimit.LimitCount),
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()

			var longBanKey = "chii:rate-limit:long:3:" + ip
			var rateLimitKey = "chii:rate-limit:rate:3:" + ip

			banned, err := script.Exec(c.Request().Context(), r, []string{longBanKey, rateLimitKey}, args).ToInt64()
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					logger.Error("failed to apply rate limit", zap.Error(err))
				}
				return err
			}

			if banned != 0 {
				return c.JSON(http.StatusTooManyRequests,
					res.Error{
						Title:       "Too Many Request",
						Description: `too many request, you have be rate limited`,
					},
				)
			}

			return next(c)
		}
	}
}
