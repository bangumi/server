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
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/config/env"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/openapi"
	"github.com/bangumi/server/web/mw/recovery"
	"github.com/bangumi/server/web/req/cf"
)

const headerProcessTime = "x-process-time-ms"
const headerServerVersion = "x-server-version"

//nolint:funlen
func New() *echo.Echo {
	app := echo.New()
	app.HTTPErrorHandler = getDefaultErrorHandler()
	app.HideBanner = true
	app.HidePort = true

	app.JSONSerializer = jsonSerializer{}

	app.IPExtractor = func(request *http.Request) string {
		ip := request.Header.Get(cf.HeaderRequestIP)
		if ip == "" {
			ra, _, _ := net.SplitHostPort(request.RemoteAddr)
			return ra
		}

		return ip
	}

	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			metrics.RequestCount.Inc()
			start := time.Now()
			c.Response().Header().Set(headerServerVersion, config.Version)

			c.Response().Before(func() {
				sub := time.Since(start)
				metrics.RequestHistogram.Observe(sub.Seconds())
				c.Response().Header().Set(headerProcessTime, strconv.FormatInt(sub.Milliseconds(), 10))
			})

			err := next(c)

			return err
		}
	})

	app.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	app.GET("/debug/pprof/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))
	app.GET("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	app.GET("/debug/pprof/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	app.GET("/debug/pprof/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	app.GET("/debug/pprof/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))

	if env.Development {
		app.Use(genFakeRequestID)
	}

	app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders: []string{headerProcessTime, headerServerVersion, cf.HeaderRequestID},
		MaxAge:        gtime.OneWeekSec,
	}))

	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request().Context()), time.Minute)
			defer cancel()

			reqID := c.Request().Header.Get(cf.HeaderRequestID)

			if reqID == "" {
				reqID = uuid.Must(uuid.NewV7()).String()
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(ctx, logger.RequestKey, &logger.RequestTrace{
				IP:    c.RealIP(),
				ReqID: reqID,
				Path:  c.Request().RequestURI,
			})))

			return next(c)
		}
	})

	app.Use(recovery.New())

	app.GET("/openapi", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/openapi/")
	})

	if env.Development {
		// fasthttp bug, it uses an internal global variable and causing data race here
		app.Static("/openapi/", "./openapi/")
		app.GET("/debug", func(c echo.Context) error {
			return c.JSON(http.StatusOK, echo.Map{
				"ip": c.RealIP(),
			})
		})
	} else {
		app.StaticFS("/openapi/", openapi.Static)
	}

	return app
}

// NewTestingApp create a base echo App for testing
// default production echo app handle panic, this doesn't.
func NewTestingApp() *echo.Echo {
	app := echo.New()
	app.HTTPErrorHandler = getDefaultErrorHandler()
	app.HideBanner = true
	app.HidePort = true

	app.Use(genFakeRequestID)

	return app
}

func Start(c config.AppConfig, app *echo.Echo) error {
	addr := c.ListenAddr()
	logger.Infoln("http server listening at", addr)
	if env.Development {
		fmt.Printf("\nvisit http://%s/\n\n", strings.ReplaceAll(addr, "0.0.0.0", "127.0.0.1"))
	}

	return errgo.Wrap(app.Start(c.ListenAddr()), "echo.Start")
}
