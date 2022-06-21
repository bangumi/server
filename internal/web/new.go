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
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/uber-go/tally/v4"
	promreporter "github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/internal/web/middleware/recovery"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/util"
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

func Start(c config.AppConfig, app *fiber.App) error {
	logger.Infoln("start http server at port", c.HTTPPort)
	addr := fmt.Sprintf(":%d", c.HTTPPort)
	if config.Development {
		addr = "127.0.0.1" + addr
	}

	return errgo.Wrap(app.Listen(addr), "fiber.App.Listen")
}

func getDefaultErrorHandler() func(*fiber.Ctx, error) error {
	var log = logger.Named("http.err").WithOptions(zap.AddStacktrace(zapcore.PanicLevel))

	return func(ctx *fiber.Ctx, err error) error {
		var e *res.HTTPError
		if errors.As(err, &e) {
			// handle expected http error
			return res.JSON(ctx.Status(e.Code), res.Error{
				Title:       utils.StatusMessage(e.Code),
				Description: e.Msg,
				Details:     util.Detail(ctx),
			})
		}

		log.Error("unexpected error",
			zap.Error(err),
			zap.String("path", ctx.Path()),
			zap.ByteString("query", ctx.Request().URI().QueryString()),
			zap.String("cf-ray", ctx.Get(req.HeaderCFRay)),
		)
		// unexpected error, return internal server error
		return res.JSON(ctx.Status(http.StatusInternalServerError), res.Error{
			Title:       "Internal Server Error",
			Description: "Unexpected Internal Server Error",
			Details:     util.DetailWithErr(ctx, err),
		})
	}
}
