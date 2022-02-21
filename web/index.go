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

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/web/middleware/recovery"
	"github.com/bangumi/server/web/res"
)

func New(scope tally.Scope, reporter promreporter.Reporter) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
		ErrorHandler:          getDefaultErrorHandler(),
		JSONEncoder: func(v interface{}) ([]byte, error) {
			//nolint:wrapcheck
			return json.MarshalIndentWithOption(v, "", "  ",
				json.DisableNormalizeUTF8(), json.DisableHTMLEscape())
		},
	})

	histogram := scope.Histogram("response_time", tally.DefaultBuckets)
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
	var errLogger = logger.Named("http.err")

	return func(c *fiber.Ctx, err error) error {
		// Default 500 status code
		code := fiber.StatusInternalServerError
		description := "Unexpected Internal Server Error"

		// router will return an un-wrapped error, so just check it like this.
		// DO NOT rewrite it to errors.Is, it's not working in this case
		if e, ok := err.(*fiber.Error); ok { //nolint:errorlint
			code = e.Code
			switch code {
			case fiber.StatusInternalServerError:
				break
			case fiber.StatusNotFound:
				description = "resource can't be found in the database or has been removed"
			default:
				description = e.Error()
			}
		} else {
			errLogger.Error(err.Error(),
				zap.String("path", c.Path()), zap.String("cf-ray", c.Get("cf-ray")))
		}

		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		return c.Status(code).JSON(res.Error{
			Title:       utils.StatusMessage(code),
			Description: description,
			Details: detail{
				Error:       err.Error(),
				Path:        c.Path(),
				QueryString: c.Request().URI().QueryArgs().String(),
			},
		})
	}
}

func detailFromRequest(c *fiber.Ctx) detail {
	return detail{
		Path:        c.Path(),
		QueryString: c.Request().URI().QueryArgs().String(),
	}
}

type detail struct {
	Error       string `json:"error,omitempty"`
	Path        string `json:"path,omitempty"`
	QueryString string `json:"query_string,omitempty"`
}
