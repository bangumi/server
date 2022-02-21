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

	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/domain"
)

// ResistRouter add all router and default 404 Handler to app.
func ResistRouter(app *fiber.App, h Handler, scope tally.Scope) {
	app.Use(newAccessUserMiddleware(h))

	reqCounter := scope.
		Tagged(map[string]string{"handler": utils.FunctionName(h.getSubject)}).
		Counter("request_count")

	app.Get("/v1/subjects/:id", func(c *fiber.Ctx) error {
		reqCounter.Inc(1)

		return c.Next()
	}, h.getSubject)

	// default 404 Handler, all router should be added before this router
	app.Use(func(c *fiber.Ctx) error {
		c.Status(fiber.StatusNotFound).
			Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)

		return c.SendString(`{
  "title": "Not Found",
  "description": "The path you requested doesn't exist",
  "detail": "This is default 404 response, if you see this response, please check your request path"
}`)
	})
}

func NewHandle(repo domain.SubjectRepo, auth domain.AuthRepo, cache cache.Generic) Handler {
	return Handler{s: repo, a: auth, cache: cache}
}

type Handler struct {
	// replace it with service, when it's too complex. Just use a repository currently.
	s     domain.SubjectRepo
	a     domain.AuthRepo
	cache cache.Generic
}
