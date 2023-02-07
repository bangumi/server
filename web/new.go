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
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic/decoder"
	"github.com/bytedance/sonic/encoder"
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

			err := next(c)

			sub := time.Since(start)
			metrics.RequestHistogram.Observe(sub.Seconds())
			c.Set(headerProcessTime, strconv.FormatInt(sub.Milliseconds(), 10))
			c.Set(headerServerVersion, config.Version)
			return err
		}
	})

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
			reqID := c.Request().Header.Get(cf.HeaderRequestID)
			reqIP := c.RealIP()

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
	app.JSONSerializer = echoJSONSerializer{}
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
