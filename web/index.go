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
	"github.com/uber-go/tally/v4"
	promreporter "github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/fx"
	"go.uber.org/zap"

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

	histogram := scope.Histogram("response_time", metrics.ResponseTimeBucket())
	app.Use(func(c *fiber.Ctx) error {
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
	var log = logger.Named("http.err")

	return func(c *fiber.Ctx, err error) error {
		log.Error("default error handler catch a error", zap.Error(err),
			zap.String("path", c.Path()), zap.String("cf-ray", c.Get("cf-ray")))

		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		return c.Status(fiber.StatusInternalServerError).JSON(res.Error{
			Title:       "Internal Server Error",
			Description: "Unexpected Internal Server Error",
			Details: util.Detail{
				Error:       err.Error(),
				Path:        c.Path(),
				QueryString: c.Request().URI().QueryArgs().String(),
			},
		})
	}
}
