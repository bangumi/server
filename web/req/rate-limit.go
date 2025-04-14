package req

import (
	"context"
	_ "embed"
	"errors"
	"net/http"

	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/web/res"
	"github.com/labstack/echo/v4"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

//go:embed ban.lua
var rateLimitLua string

func RateLimit(r rueidis.Client) echo.MiddlewareFunc {
	script := rueidis.NewLuaScript(rateLimitLua)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()

			var longBanKey = "chii-rate-limit:long:" + ip
			var rateLimitKey = "chii-rate-limit:rate:" + ip

			banned, err := script.Exec(c.Request().Context(), r, []string{longBanKey, rateLimitKey}, nil).ToInt64()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				logger.Error("failed to apply rate limit", zap.Error(err))
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
