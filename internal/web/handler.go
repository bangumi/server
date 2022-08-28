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
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/uber-go/tally/v4"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/web/frontend"
	"github.com/bangumi/server/internal/web/handler"
	"github.com/bangumi/server/internal/web/handler/character"
	"github.com/bangumi/server/internal/web/handler/person"
	"github.com/bangumi/server/internal/web/handler/subject"
	"github.com/bangumi/server/internal/web/handler/user"
	"github.com/bangumi/server/internal/web/middleware/origin"
	"github.com/bangumi/server/internal/web/middleware/referer"
	"github.com/bangumi/server/internal/web/middleware/ua"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/util"
)

// AddRouters add all router and default 404 Handler to app.
//
//nolint:funlen
func AddRouters(
	app *fiber.App,
	c config.AppConfig,
	h handler.Handler,
	scope tally.Scope,
	userHandler user.User,
	personHandler person.Person,
	characterHandler character.Character,
	subjectHandler subject.Subject,
) {
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

	v0 := app.Group("/v0/", h.MiddlewareAccessTokenAuth)

	v0.Post("/search/subjects", h.Search)

	v0.Get("/subjects/:id", addMetrics(subjectHandler.Get))
	v0.Get("/subjects/:id/image", addMetrics(subjectHandler.GetImage))
	v0.Get("/subjects/:id/persons", addMetrics(subjectHandler.GetRelatedPersons))
	v0.Get("/subjects/:id/subjects", addMetrics(subjectHandler.GetRelatedSubjects))
	v0.Get("/subjects/:id/characters", addMetrics(subjectHandler.GetRelatedCharacters))
	v0.Get("/persons/:id", addMetrics(personHandler.Get))
	v0.Get("/persons/:id/image", addMetrics(personHandler.GetImage))
	v0.Get("/persons/:id/subjects", addMetrics(personHandler.GetRelatedSubjects))
	v0.Get("/persons/:id/characters", addMetrics(personHandler.GetRelatedCharacters))
	v0.Get("/characters/:id", addMetrics(characterHandler.Get))
	v0.Get("/characters/:id/image", addMetrics(characterHandler.GetImage))
	v0.Get("/characters/:id/subjects", addMetrics(characterHandler.GetRelatedSubjects))
	v0.Get("/characters/:id/persons", addMetrics(characterHandler.GetRelatedPersons))
	v0.Get("/episodes/:id", addMetrics(h.GetEpisode))
	v0.Get("/episodes", addMetrics(h.ListEpisode))

	v0.Get("/me", addMetrics(userHandler.GetCurrent))
	v0.Get("/users/:username", addMetrics(userHandler.Get))
	v0.Get("/users/:username/collections", addMetrics(userHandler.ListSubjectCollection))
	v0.Get("/users/:username/collections/:subject_id", addMetrics(userHandler.GetSubjectCollection))
	v0.Get("/users/-/collections/-/episodes/:episode_id", h.NeedLogin, addMetrics(userHandler.GetEpisodeCollection))
	v0.Get("/users/-/collections/:subject_id/episodes", h.NeedLogin, addMetrics(userHandler.GetSubjectEpisodeCollection))
	v0.Patch("/users/-/collections/:subject_id", req.JSON, h.NeedLogin, addMetrics(userHandler.PatchSubjectCollection))
	v0.Patch("/users/-/collections/:subject_id/episodes",
		req.JSON, h.NeedLogin,
		addMetrics(userHandler.PatchEpisodeCollectionBatch),
	)
	v0.Get("/users/:username/avatar", addMetrics(userHandler.GetAvatar))

	v0.Get("/indices/:id", addMetrics(h.GetIndex))
	v0.Get("/indices/:id/subjects", addMetrics(h.GetIndexSubjects))

	v0.Get("/revisions/persons/:id", addMetrics(h.GetPersonRevision))
	v0.Get("/revisions/persons", addMetrics(h.ListPersonRevision))
	v0.Get("/revisions/subjects/:id", addMetrics(h.GetSubjectRevision))
	v0.Get("/revisions/subjects", addMetrics(h.ListSubjectRevision))
	v0.Get("/revisions/characters/:id", addMetrics(h.GetCharacterRevision))
	v0.Get("/revisions/characters", addMetrics(h.ListCharacterRevision))

	var originMiddleware = origin.New(fmt.Sprintf("https://%s", c.WebDomain))
	var refererMiddleware = referer.New(fmt.Sprintf("https://%s/", c.WebDomain))

	var CORSBlockMiddleware []fiber.Handler
	if c.WebDomain != "" {
		CORSBlockMiddleware = []fiber.Handler{originMiddleware, refererMiddleware}
	}

	// frontend private api
	private := app.Group("/p/", append(CORSBlockMiddleware, h.MiddlewareSessionAuth)...)

	private.Post("/login", req.JSON, addMetrics(h.PrivateLogin))
	private.Post("/logout", addMetrics(h.PrivateLogout))
	private.Get("/me", addMetrics(userHandler.GetCurrent))
	private.Get("/groups/:name", addMetrics(h.GetGroupProfileByNamePrivate))
	private.Get("/groups/:name/members", addMetrics(h.ListGroupMembersPrivate))

	private.Get("/groups/:name/topics", addMetrics(h.ListGroupTopics))
	private.Get("/subjects/:id/topics", addMetrics(h.ListSubjectTopics))

	private.Get("/groups/-/topics/:topic_id", addMetrics(h.GetGroupTopic))
	private.Get("/subjects/:id/topics/:topic_id", addMetrics(h.GetSubjectTopic))
	private.Get("/indices/:id/comments", addMetrics(h.GetIndexComments))
	private.Get("/episodes/:id/comments", addMetrics(h.GetEpisodeComments))
	private.Get("/characters/:id/comments", addMetrics(h.GetCharacterComments))
	private.Get("/persons/:id/comments", addMetrics(h.GetPersonComments))

	// un-documented
	private.Post("/access-tokens", req.JSON, addMetrics(h.CreatePersonalAccessToken))
	private.Delete("/access-tokens", req.JSON, addMetrics(h.DeletePersonalAccessToken))

	if c.WebDomain != "" {
		CORSBlockMiddleware = []fiber.Handler{originMiddleware}
	}

	privateHTML := app.Group("/demo/", append(CORSBlockMiddleware, h.MiddlewareSessionAuth)...)
	privateHTML.Get("/login", addMetrics(h.PageLogin))
	privateHTML.Get("/access-token", addMetrics(h.PageListAccessToken))
	privateHTML.Get("/access-token/create", addMetrics(h.PageCreateAccessToken))

	app.Use("/static/", filesystem.New(filesystem.Config{
		PathPrefix: "static",
		Root:       http.FS(frontend.StaticFS),
		MaxAge:     gtime.OneWeekSec,
	}))

	// default 404 Handler, all router should be added before this router
	app.Use(func(c *fiber.Ctx) error {
		return res.JSON(c.Status(http.StatusNotFound), res.Error{
			Title:       "Not Found",
			Description: "This is default response, if you see this response, please check your request",
			Details:     util.Detail(c),
		})
	})
}

func trimFuncName(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, "github.com/bangumi/server/internal/"), "-fm")
}
