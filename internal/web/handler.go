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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/config"
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
	userHandler user.User,
	personHandler person.Person,
	characterHandler character.Character,
	subjectHandler subject.Subject,
) {
	app.Get("/", indexPage)

	app.Use(ua.DisableDefaultHTTPLibrary)

	v0 := app.Group("/v0/", h.MiddlewareAccessTokenAuth)

	v0.Post("/search/subjects", h.Search)

	v0.Get("/subjects/:id", subjectHandler.Get)
	v0.Get("/subjects/:id/image", subjectHandler.GetImage)
	v0.Get("/subjects/:id/persons", subjectHandler.GetRelatedPersons)
	v0.Get("/subjects/:id/subjects", subjectHandler.GetRelatedSubjects)
	v0.Get("/subjects/:id/characters", subjectHandler.GetRelatedCharacters)
	v0.Get("/persons/:id", personHandler.Get)
	v0.Get("/persons/:id/image", personHandler.GetImage)
	v0.Get("/persons/:id/subjects", personHandler.GetRelatedSubjects)
	v0.Get("/persons/:id/characters", personHandler.GetRelatedCharacters)
	v0.Get("/characters/:id", characterHandler.Get)
	v0.Get("/characters/:id/image", characterHandler.GetImage)
	v0.Get("/characters/:id/subjects", characterHandler.GetRelatedSubjects)
	v0.Get("/characters/:id/persons", characterHandler.GetRelatedPersons)
	v0.Get("/episodes/:id", h.GetEpisode)
	v0.Get("/episodes", h.ListEpisode)

	v0.Get("/me", userHandler.GetCurrent)
	v0.Get("/users/:username", userHandler.Get)
	v0.Get("/users/:username/collections", userHandler.ListSubjectCollection)
	v0.Get("/users/:username/collections/:subject_id", userHandler.GetSubjectCollection)
	v0.Get("/users/-/collections/-/episodes/:episode_id", h.NeedLogin, userHandler.GetEpisodeCollection)
	v0.Put("/users/-/collections/-/episodes/:episode_id", req.JSON, h.NeedLogin, userHandler.PutEpisodeCollection)
	v0.Get("/users/-/collections/:subject_id/episodes", h.NeedLogin, userHandler.GetSubjectEpisodeCollection)
	v0.Patch("/users/-/collections/:subject_id", req.JSON, h.NeedLogin, userHandler.PatchSubjectCollection)
	v0.Patch("/users/-/collections/:subject_id/episodes",
		req.JSON, h.NeedLogin, userHandler.PatchEpisodeCollectionBatch,
	)
	v0.Get("/users/:username/avatar", userHandler.GetAvatar)

	v0.Get("/indices/:id", h.GetIndex)
	v0.Get("/indices/:id/subjects", h.GetIndexSubjects)
	// indices
	v0.Post("/indices", req.JSON, h.NeedLogin, h.NewIndex)
	v0.Put("/indices/:id", req.JSON, h.NeedLogin, h.UpdateIndex)
	// indices subjects
	v0.Post("/indices/:id/subjects", req.JSON, h.NeedLogin, h.AddIndexSubject)
	v0.Put("/indices/:id/subjects/:subject_id", req.JSON, h.NeedLogin, h.UpdateIndexSubject)
	v0.Delete("/indices/:id/subjects/:subject_id", h.NeedLogin, h.RemoveIndexSubject)

	v0.Get("/revisions/persons/:id", h.GetPersonRevision)
	v0.Get("/revisions/persons", h.ListPersonRevision)
	v0.Get("/revisions/subjects/:id", h.GetSubjectRevision)
	v0.Get("/revisions/subjects", h.ListSubjectRevision)
	v0.Get("/revisions/characters/:id", h.GetCharacterRevision)
	v0.Get("/revisions/characters", h.ListCharacterRevision)

	var originMiddleware = origin.New(fmt.Sprintf("https://%s", c.WebDomain))
	var refererMiddleware = referer.New(fmt.Sprintf("https://%s/", c.WebDomain))

	var CORSBlockMiddleware []fiber.Handler
	if c.WebDomain != "" {
		CORSBlockMiddleware = []fiber.Handler{originMiddleware, refererMiddleware}
	}

	// frontend private api
	private := app.Group("/p/", append(CORSBlockMiddleware, h.MiddlewareSessionAuth)...)

	private.Post("/login", req.JSON, h.PrivateLogin)
	private.Post("/logout", h.PrivateLogout)
	private.Get("/me", userHandler.GetCurrent)
	private.Get("/groups/:name", h.GetGroupProfileByNamePrivate)
	private.Get("/groups/:name/members", h.ListGroupMembersPrivate)

	private.Get("/groups/:name/topics", h.ListGroupTopics)
	private.Get("/subjects/:id/topics", h.ListSubjectTopics)

	private.Get("/groups/-/topics/:topic_id", h.GetGroupTopic)
	private.Get("/subjects/:id/topics/:topic_id", h.GetSubjectTopic)
	private.Get("/indices/:id/comments", h.GetIndexComments)
	private.Get("/episodes/:id/comments", h.GetEpisodeComments)
	private.Get("/characters/:id/comments", h.GetCharacterComments)
	private.Get("/persons/:id/comments", h.GetPersonComments)

	// un-documented
	private.Post("/access-tokens", req.JSON, h.CreatePersonalAccessToken)
	private.Delete("/access-tokens", req.JSON, h.DeletePersonalAccessToken)

	if c.WebDomain != "" {
		CORSBlockMiddleware = []fiber.Handler{originMiddleware}
	}

	privateHTML := app.Group("/demo/", append(CORSBlockMiddleware, h.MiddlewareSessionAuth)...)
	privateHTML.Get("/login", h.PageLogin)
	privateHTML.Get("/access-token", h.PageListAccessToken)
	privateHTML.Get("/access-token/create", h.PageCreateAccessToken)

	// default 404 Handler, all router should be added before this router
	app.Use(func(c *fiber.Ctx) error {
		return res.JSON(c.Status(http.StatusNotFound), res.Error{
			Title:       "Not Found",
			Description: "This is default response, if you see this response, please check your request",
			Details:     util.Detail(c),
		})
	})
}
