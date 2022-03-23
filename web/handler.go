// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
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

	"github.com/bangumi/server/web/handler"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/util"
)

// ResistRouter add all router and default 404 Handler to app.
func ResistRouter(app *fiber.App, h handler.Handler, scope tally.Scope) {
	app.Use(h.MiddlewareAccessUser())

	// add logger wrapper and metrics counter
	addMetrics := func(handler fiber.Handler) fiber.Handler {
		reqCounter := scope.
			Tagged(map[string]string{"handler": utils.FunctionName(handler)}).
			Counter("request_count")

		return func(ctx *fiber.Ctx) error {
			reqCounter.Inc(1)
			return handler(ctx)
		}
	}

	app.Get("/v0/subjects/:id", addMetrics(h.GetSubject))
	app.Get("/v0/subjects/:id/persons", addMetrics(h.GetSubjectRelatedPersons))
	app.Get("/v0/subjects/:id/subjects", addMetrics(h.GetSubjectRelatedSubjects))
	app.Get("/v0/subjects/:id/characters", addMetrics(h.GetSubjectRelatedCharacters))
	app.Get("/v0/persons/:id", addMetrics(h.GetPerson))
	app.Get("/v0/persons/:id/subjects", addMetrics(h.GetPersonRelatedSubjects))
	app.Get("/v0/persons/:id/characters", addMetrics(h.GetPersonRelatedCharacters))
	app.Get("/v0/characters/:id", addMetrics(h.GetCharacter))
	app.Get("/v0/characters/:id/subjects", addMetrics(h.GetCharacterRelatedSubjects))
	app.Get("/v0/characters/:id/persons", addMetrics(h.GetCharacterRelatedPersons))
	app.Get("/v0/episodes/:id", addMetrics(h.GetEpisode))
	app.Get("/v0/episodes", addMetrics(h.ListEpisode))
	app.Get("/v0/me", addMetrics(h.GetCurrentUser))
	app.Get("/v0/users/:username/collections", addMetrics(h.ListCollection))
	app.Get("/v0/indices/:id", addMetrics(h.GetIndex))
	app.Get("/v0/indices/:id/subjects", addMetrics(h.GetIndexSubjects))

	app.Get("/v0/revisions/persons/:id", addMetrics(h.GetPersonRevision))
	app.Get("/v0/revisions/persons", addMetrics(h.ListPersonRevision))
	app.Get("/v0/revisions/subjects/:id", addMetrics(h.GetSubjectRevision))
	app.Get("/v0/revisions/subjects", addMetrics(h.ListSubjectRevision))

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
