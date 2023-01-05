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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic/decoder"
	"github.com/bytedance/sonic/encoder"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/config/env"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/random"
	"github.com/bangumi/server/openapi"
	"github.com/bangumi/server/web/middleware/recovery"
	"github.com/bangumi/server/web/req/cf"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/util"
)

const headerProcessTime = "x-process-time-ms"
const headerServerVersion = "x-server-version"

type echoJSONSerializer struct {
}

func (e echoJSONSerializer) Serialize(c echo.Context, i any, indent string) error {
	enc := encoder.NewStreamEncoder(c.Response())

	enc.SetIndent("", indent)

	return enc.Encode(i) //nolint:wrapcheck
}

func (e echoJSONSerializer) Deserialize(c echo.Context, i any) error {
	return decoder.NewStreamDecoder(c.Request().Body).Decode(i) //nolint:wrapcheck
}

var _ echo.JSONSerializer = echoJSONSerializer{}

//nolint:funlen
func New() *echo.Echo {
	app := echo.New()
	app.JSONSerializer = echoJSONSerializer{}
	app.HTTPErrorHandler = getDefaultErrorHandler()
	app.HideBanner = true
	app.HidePort = true

	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			metrics.RequestCount.Inc()
			start := time.Now()

			err := next(c)

			sub := time.Since(start)
			metrics.RequestHistogram.Observe(sub.Seconds())
			c.Set(headerProcessTime, strconv.FormatInt(sub.Milliseconds(), 10))
			c.Set(headerServerVersion, config.Version)
			return err
		}
	})

	if env.Development {
		app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				devRequestID := "fake-ray-" + random.Base62String(10)
				c.Request().Header.Set(cf.HeaderRequestID, devRequestID)
				c.Request().Header.Set(cf.HeaderRequestIP, c.Request().RemoteAddr)
				c.Set(cf.HeaderRequestID, devRequestID)

				return next(c)
			}
		})
	}

	app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders: []string{headerProcessTime, headerServerVersion, cf.HeaderRequestID},
		MaxAge:        gtime.OneWeekSec,
	}))

	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get(cf.HeaderRequestID)
			reqIP := c.Request().Header.Get(cf.HeaderRequestIP)

			c.SetRequest(c.Request().
				WithContext(context.WithValue(context.Background(), logger.RequestKey, &logger.RequestTrace{
					IP:    reqIP,
					ReqID: reqID,
				})))

			return next(c)
		}
	})

	app.Use(recovery.New())

	app.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	addProfile(app)

	app.GET("/openapi", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/openapi/")
	})

	if env.Development {
		// fasthttp bug, it uses an internal global variable and causing data race here
		app.Static("/openapi/", "./openapi/")
	} else {
		app.StaticFS("/openapi/", openapi.Static)
	}

	return app
}

func addProfile(app *echo.Echo) {
	app.GET("/debug/pprof/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))
	app.GET("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	app.GET("/debug/pprof/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	app.GET("/debug/pprof/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	app.Any("/debug/pprof/", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
}

func Start(c config.AppConfig, app *echo.Echo) error {
	addr := c.ListenAddr()
	logger.Infoln("http server listening at", addr)
	if env.Development {
		fmt.Printf("\nvisit http://%s/\n\n", strings.ReplaceAll(addr, "0.0.0.0", "127.0.0.1"))
	}

	return errgo.Wrap(app.Start(c.ListenAddr()), "echo.Start")
}

func getDefaultErrorHandler() echo.HTTPErrorHandler {
	var log = logger.Named("http.err").
		WithOptions(zap.AddStacktrace(zapcore.PanicLevel), zap.WithCaller(false))

	return func(err error, c echo.Context) {
		{
			var e res.HTTPError
			if errors.As(err, &e) {
				// handle expected http error
				_ = c.JSON(e.Code, res.Error{
					Title:       http.StatusText(e.Code),
					Description: e.Msg,
					Details:     util.Detail(c),
				})
				return
			}
		}

		{
			//nolint:forbidigo,errorlint
			if e, ok := err.(*echo.HTTPError); ok {
				log.Error("unexpected echo error",
					zap.Int("code", e.Code),
					zap.Any("message", e.Message),
					zap.String("path", c.Request().URL.Path),
					zap.String("query", c.Request().URL.RawQuery),
					zap.String("cf-ray", c.Request().Header.Get(cf.HeaderRequestID)),
				)

				_ = c.JSON(http.StatusInternalServerError, res.Error{
					Title:       http.StatusText(e.Code),
					Description: e.Error(),
					Details:     util.DetailWithErr(c, err),
				})
				return
			}
		}

		log.Error("unexpected error",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("query", c.Request().URL.RawQuery),
			zap.String("cf-ray", c.Request().Header.Get(cf.HeaderRequestID)),
		)

		// unexpected error, return internal server error
		_ = res.InternalError(c, err, "Unexpected Internal Server Error")
	}
}
