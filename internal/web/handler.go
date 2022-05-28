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
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gookit/goutil/timex"
	"github.com/uber-go/tally/v4"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/web/frontend"
	"github.com/bangumi/server/internal/web/handler"
	"github.com/bangumi/server/internal/web/middleware/origin"
	"github.com/bangumi/server/internal/web/middleware/referer"
	"github.com/bangumi/server/internal/web/middleware/ua"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/res/code"
	"github.com/bangumi/server/internal/web/util"
)

// ResistRouter add all router and default 404 Handler to app.
//nolint:funlen
func ResistRouter(app *fiber.App, h handler.Handler, scope tally.Scope) {
	app.Use(ua.DisableDefaultHTTPLibrary)

	// add logger wrapper and metrics counter
	addMetrics := func(handler fiber.Handler) fiber.Handler {
		reqCounter := scope.
			Tagged(map[string]string{"handler": trimFuncName(utils.FunctionName(handler))}).
			Counter("request_count")

		return func(ctx *fiber.Ctx) error {
			reqCounter.Inc(1)
			return handler(ctx)
		}
	}

	v0 := app.Group("/v0/", h.AccessTokenAuthMiddleware)

	v0.Get("/subjects/:id", addMetrics(h.GetSubject))
	v0.Get("/subjects/:id/persons", addMetrics(h.GetSubjectRelatedPersons))
	v0.Get("/subjects/:id/subjects", addMetrics(h.GetSubjectRelatedSubjects))
	v0.Get("/subjects/:id/characters", addMetrics(h.GetSubjectRelatedCharacters))
	v0.Get("/persons/:id", addMetrics(h.GetPerson))
	v0.Get("/persons/:id/subjects", addMetrics(h.GetPersonRelatedSubjects))
	v0.Get("/persons/:id/characters", addMetrics(h.GetPersonRelatedCharacters))
	v0.Get("/characters/:id", addMetrics(h.GetCharacter))
	v0.Get("/characters/:id/subjects", addMetrics(h.GetCharacterRelatedSubjects))
	v0.Get("/characters/:id/persons", addMetrics(h.GetCharacterRelatedPersons))
	v0.Get("/episodes/:id", addMetrics(h.GetEpisode))
	v0.Get("/episodes", addMetrics(h.ListEpisode))

	v0.Get("/me", addMetrics(h.GetCurrentUser))
	v0.Get("/users/:username/collections", addMetrics(h.ListCollection))
	v0.Get("/users/:username/collections/:subject_id", addMetrics(h.GetCollection))
	v0.Get("/users/:username", addMetrics(h.GetUser))

	v0.Get("/indices/:id", addMetrics(h.GetIndex))
	v0.Get("/indices/:id/subjects", addMetrics(h.GetIndexSubjects))

	v0.Get("/revisions/persons/:id", addMetrics(h.GetPersonRevision))
	v0.Get("/revisions/persons", addMetrics(h.ListPersonRevision))
	v0.Get("/revisions/subjects/:id", addMetrics(h.GetSubjectRevision))
	v0.Get("/revisions/subjects", addMetrics(h.ListSubjectRevision))
	v0.Get("/revisions/characters/:id", addMetrics(h.GetCharacterRevision))
	v0.Get("/revisions/characters", addMetrics(h.ListCharacterRevision))

	app.Post("/_private/revoke", req.JSON, addMetrics(h.RevokeSession))

	// frontend private api
	private := app.Group("/p/",
		origin.New(config.FrontendOrigin), referer.New(config.FrontendOrigin+"/"), h.SessionAuthMiddleware)

	private.Post("/login", req.JSON, addMetrics(h.PrivateLogin))
	private.Post("/logout", addMetrics(h.PrivateLogout))

	private.Get("/me", addMetrics(h.GetCurrentUser))

	privateHTML := app.Group("/demo/", origin.New(config.FrontendOrigin), h.SessionAuthMiddleware)
	privateHTML.Get("/login", h.PageLogin)

	app.Use("/static/", filesystem.New(filesystem.Config{
		PathPrefix: "static",
		Root:       http.FS(frontend.StaticFS),
		MaxAge:     timex.OneWeekSec,
	}))

	// default 404 Handler, all router should be added before this router
	app.Use(func(c *fiber.Ctx) error {
		return res.JSON(c.Status(code.NotFound), res.Error{
			Title:       "Not Found",
			Description: "This is default response, if you see this response, please check your request path",
			Details:     util.DetailFromRequest(c),
		})
	})
}

func trimFuncName(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, "github.com/bangumi/server/web/handler.Handler."), "-fm")
}
