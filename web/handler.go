// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
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
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/web/handler"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/util"
)

// ResistRouter add all router and default 404 Handler to app.
func ResistRouter(app *fiber.App, h handler.Handler, scope tally.Scope) {
	app.Use(h.MiddlewareAccessUser())
	var log = logger.Named("http.err")

	// add logger wrapper and metrics counter
	addHandle := func(
		reg func(path string, handlers ...fiber.Handler) fiber.Router,
		path string,
		handler fiber.Handler,
	) {
		reqCounter := scope.
			Tagged(map[string]string{"handler": utils.FunctionName(handler)}).
			Counter("request_count")

		middle := func(c *fiber.Ctx) error {
			reqCounter.Inc(1)
			return c.Next()
		}

		reg(path, middle, handlerWrapper(log, handler))
	}

	addHandle(app.Get, "/v0/subjects/:id", h.GetSubject)
	addHandle(app.Get, "/v0/subjects/:id/persons", h.GetSubjectRelatedPersons)
	addHandle(app.Get, "/v0/subjects/:id/subjects", h.GetSubjectRelatedSubjects)
	addHandle(app.Get, "/v0/subjects/:id/characters", h.GetSubjectRelatedCharacters)

	addHandle(app.Get, "/v0/persons/:id", h.GetPerson)
	addHandle(app.Get, "/v0/persons/:id/subjects", h.GetPersonRelatedSubjects)
	addHandle(app.Get, "/v0/persons/:id/characters", h.GetPersonRelatedCharacters)

	addHandle(app.Get, "/v0/characters/:id", h.GetCharacter)
	addHandle(app.Get, "/v0/characters/:id/subjects", h.GetCharacterRelatedSubjects)
	addHandle(app.Get, "/v0/characters/:id/persons", h.GetCharacterRelatedPersons)

	addHandle(app.Get, "/v0/episodes/:id", h.GetEpisode)
	addHandle(app.Get, "/v0/episodes", h.ListEpisode)

	addHandle(app.Get, "/v0/me", h.GetCurrentUser)
	addHandle(app.Get, "/v0/users/:username/collections", h.ListCollection)

	addHandle(app.Get, "/v0/indices/:id", h.GetIndex)
	addHandle(app.Get, "/v0/indices/:id/subjects", h.GetIndexSubjects)

	// default 404 Handler, all router should be added before this router
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(res.Error{
			Title: "Not Found",
			Description: "This is default response, " +
				"if you see this response, please check your request path",
			Details: util.DetailFromRequest(c),
		})
	})
}

// wrap handler to make trace stack works.
func handlerWrapper(log *zap.Logger, handler fiber.Handler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		err := handler(ctx)
		if err == nil {
			return nil
		}
		// Default 500 status code
		code := fiber.StatusInternalServerError
		description := "Unexpected Internal Server Error"

		// router will return an un-wrapped error, so just check it like this.
		// DO NOT rewrite it to errors.Is.
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
			log.Error("unexpected error", zap.Error(err),
				zap.String("path", ctx.Path()), zap.String("cf-ray", ctx.Get("cf-ray")))
		}

		ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		return ctx.Status(code).JSON(res.Error{
			Title:       utils.StatusMessage(code),
			Description: description,
			Details: util.Detail{
				Error:       err.Error(),
				Path:        ctx.Path(),
				QueryString: utils.UnsafeString(ctx.Request().URI().QueryString()),
			},
		})
	}
}
