// Copyright (c) 2021-2022 Trim21 <trim21.me@gmail.com>
//
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
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/uber-go/tally/v4"
	promreporter "github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/web/middleware/recovery"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/util"
)

func New(scope tally.Scope, reporter promreporter.Reporter) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
		ErrorHandler:          getDefaultErrorHandler(),
		JSONEncoder:           json.MarshalNoEscape,
	})

	count := scope.Counter("request_count_total")
	histogram := scope.Histogram("response_time", metrics.ResponseTimeBucket())
	app.Use(func(c *fiber.Ctx) error {
		count.Inc(1)
		start := time.Now()

		err := c.Next()

		sub := time.Since(start)
		histogram.RecordDuration(sub)
		c.Set("x-process-time-ms", strconv.FormatInt(sub.Milliseconds(), 10))
		c.Set("x-server-version", config.Version)

		return err
	})

	app.Use(recovery.New())
	app.Get("/metrics", adaptor.HTTPHandler(reporter.HTTPHandler()))

	return app
}

func Listen(lc fx.Lifecycle, c config.AppConfig, app *fiber.App) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Infoln("start http server at port", c.HTTPPort)
			addr := fmt.Sprintf(":%d", c.HTTPPort)
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return errgo.Wrap(err, "failed to start listener")
			}

			go func() {
				err := app.Listener(ln)
				if err != nil {
					logger.Panic("failed to start fiber app on listener", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return errgo.Wrap(app.Shutdown(), "web.Shutdown")
		},
	})
}

func getDefaultErrorHandler() func(c *fiber.Ctx, err error) error {
	var log = logger.Named("http.err").WithOptions(zap.AddStacktrace(zapcore.PanicLevel))

	return func(ctx *fiber.Ctx, err error) error {
		// Default 500 status code
		code := fiber.StatusInternalServerError
		title := "Internal Server Error"
		description := "Unexpected Internal Server Error"

		// router will return an un-wrapped error, so just check it like this.
		// DO NOT rewrite it to errors.Is.
		if e, ok := err.(*fiber.Error); ok { //nolint:errorlint
			code = e.Code
			switch e.Code {
			case fiber.StatusInternalServerError:
				break
			case fiber.StatusNotFound:
				description = "resource can't be found in the database or has been removed"
				title = utils.StatusMessage(code)
			default:
				description = e.Error()
				title = utils.StatusMessage(code)
			}
		} else {
			log.Error("unexpected error",
				zap.Error(err),
				zap.String("path", ctx.Path()),
				zap.ByteString("query", ctx.Request().URI().QueryString()),
				zap.String("cf-ray", ctx.Get("cf-ray")),
			)
		}

		return ctx.Status(code).JSON(res.Error{
			Title:       title,
			Description: description,
			Details: util.Detail{
				Error:       err.Error(),
				Path:        ctx.Path(),
				QueryString: utils.UnsafeString(ctx.Request().URI().QueryString()),
			},
		})
	}
}
