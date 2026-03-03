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
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/web/req/cf"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/util"
)

func globalNotFoundHandler(c echo.Context) error {
	return c.JSON(http.StatusNotFound, res.Error{
		Title:       "Not Found",
		Description: "This is default response, if you see this response, please check your request",
		Details:     util.Detail(c),
	})
}

//nolint:funlen
func getDefaultErrorHandler() echo.HTTPErrorHandler {
	var log = logger.Named("http.err").
		WithOptions(zap.AddStacktrace(zapcore.PanicLevel), zap.WithCaller(false))

	return func(err error, c echo.Context) {
		reqID := c.Request().Header.Get(cf.HeaderRequestID)

		{
			var e res.HTTPError
			if errors.As(err, &e) {
				// handle expected http error
				_ = c.JSON(e.Code, res.Error{
					Title:       http.StatusText(e.Code),
					Description: e.Msg,
					RequestID:   reqID,
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
					zap.String("request_method", c.Request().Method),
					zap.String("request_uri", c.Request().URL.Path),
					zap.String("request_query", c.Request().URL.RawQuery),
					zap.String("request_id", reqID),
				)

				_ = c.JSON(http.StatusInternalServerError, res.Error{
					Title:       http.StatusText(e.Code),
					Description: e.Error(),
					RequestID:   reqID,
					Details:     util.DetailWithErr(c, err),
				})
				return
			}
		}

		if errors.Is(err, context.Canceled) {
			_ = c.NoContent(http.StatusNoContent)
			return
		}

		log.Error("unexpected error",
			zap.Error(err),
			zap.String("request_method", c.Request().Method),
			zap.String("request_uri", c.Request().URL.Path),
			zap.String("request_query", c.Request().URL.RawQuery),
			zap.String("request_id", reqID),
		)

		// unexpected error, return internal server error
		_ = res.InternalError(c, err, "Unexpected Internal Server Error")
	}
}
