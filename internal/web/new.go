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
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/config/env"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/random"
	"github.com/bangumi/server/internal/web/middleware/recovery"
	"github.com/bangumi/server/internal/web/req/cf"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/util"
)

const headerProcessTime = "x-process-time-ms"
const headerServerVersion = "x-server-version"

func New() *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
		ErrorHandler:          getDefaultErrorHandler(),
		JSONEncoder:           sonic.Marshal,
	})

	app.Use(func(c *fiber.Ctx) error {
		metrics.RequestCount.Inc()
		start := time.Now()

		err := c.Next()

		sub := time.Since(start)
		metrics.RequestHistogram.Observe(sub.Seconds())
		c.Set(headerProcessTime, strconv.FormatInt(sub.Milliseconds(), 10))
		c.Set(headerServerVersion, config.Version)
		return err
	})

	if env.Development {
		app.Use(func(c *fiber.Ctx) error {
			devRequestID := "fake-ray-" + random.Base62String(10)
			c.Request().Header.Set(cf.HeaderRequestID, devRequestID)
			c.Request().Header.Set(cf.HeaderRequestIP, c.Context().RemoteIP().String())
			c.Set(cf.HeaderRequestID, devRequestID)

			return c.Next()
		})
		app.Use(cors.New(cors.Config{
			MaxAge:        gtime.OneWeekSec,
			ExposeHeaders: strings.Join([]string{headerProcessTime, headerServerVersion, cf.HeaderRequestID}, ","),
		}))
	}

	app.Use(recovery.New())
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	addProfile(app)

	if env.Development {
		app.Use("/openapi/", openapiHandler)
	}

	return app
}

func addProfile(app *fiber.App) {
	app.Get("/debug/pprof/cmdline", adaptor.HTTPHandlerFunc(pprof.Cmdline))
	app.Get("/debug/pprof/profile", adaptor.HTTPHandlerFunc(pprof.Profile))
	app.Get("/debug/pprof/symbol", adaptor.HTTPHandlerFunc(pprof.Symbol))
	app.Get("/debug/pprof/trace", adaptor.HTTPHandlerFunc(pprof.Trace))
	app.Use("/debug/pprof/", adaptor.HTTPHandlerFunc(pprof.Index))
}

func Start(c config.AppConfig, app *fiber.App) error {
	addr := c.ListenAddr()
	logger.Infoln("http server listening at", addr)
	if env.Development {
		fmt.Printf("\nvisit http://%s/\n\n", strings.ReplaceAll(addr, "0.0.0.0", "127.0.0.1"))
	}

	return errgo.Wrap(app.Listen(c.ListenAddr()), "fiber.App.Listen")
}

func getDefaultErrorHandler() func(*fiber.Ctx, error) error {
	var log = logger.Named("http.err").
		WithOptions(zap.AddStacktrace(zapcore.PanicLevel), zap.WithCaller(false))

	return func(ctx *fiber.Ctx, err error) error {
		var e res.HTTPError
		if errors.As(err, &e) {
			// handle expected http error
			return res.JSON(ctx.Status(e.Code), res.Error{
				Title:       utils.StatusMessage(e.Code),
				Description: e.Msg,
				Details:     util.Detail(ctx),
			})
		}

		//nolint:forbidigo,errorlint
		if fErr, ok := err.(*fiber.Error); ok {
			log.Error("unexpected fiber error",
				zap.Int("code", fErr.Code),
				zap.String("message", fErr.Message),
				zap.String("path", ctx.Path()),
				zap.ByteString("query", ctx.Request().URI().QueryString()),
				zap.String("cf-ray", ctx.Get(cf.HeaderRequestID)),
			)

			return res.JSON(ctx.Status(http.StatusInternalServerError), res.Error{
				Title:       utils.StatusMessage(fErr.Code),
				Description: fErr.Message,
				Details:     util.DetailWithErr(ctx, err),
			})
		}

		log.Error("unexpected error",
			zap.Error(err),
			zap.String("path", ctx.Path()),
			zap.ByteString("query", ctx.Request().URI().QueryString()),
			zap.String("cf-ray", ctx.Get(cf.HeaderRequestID)),
		)

		// unexpected error, return internal server error
		return res.InternalError(ctx, err, "Unexpected Internal Server Error")
	}
}
