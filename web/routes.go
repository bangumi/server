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

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/handler"
	"github.com/bangumi/server/web/handler/character"
	"github.com/bangumi/server/web/handler/common"
	"github.com/bangumi/server/web/handler/index"
	"github.com/bangumi/server/web/handler/notification"
	"github.com/bangumi/server/web/handler/person"
	"github.com/bangumi/server/web/handler/pm"
	"github.com/bangumi/server/web/handler/subject"
	"github.com/bangumi/server/web/handler/user"
	"github.com/bangumi/server/web/middleware/origin"
	"github.com/bangumi/server/web/middleware/referer"
	"github.com/bangumi/server/web/middleware/ua"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/util"
)

// AddRouters add all router and default 404 Handler to app.
//
//nolint:funlen
func AddRouters(
	app *echo.Echo,
	c config.AppConfig,
	common common.Common,
	h handler.Handler,
	userHandler user.User,
	personHandler person.Person,
	characterHandler character.Character,
	pmHandler pm.PrivateMessage,
	notificationHandler notification.Notification,
	subjectHandler subject.Subject,
	indexHandler index.Handler,
) {
	app.GET("/", indexPage())

	app.Use(ua.DisableDefaultHTTPLibrary)

	v0 := app.Group("/v0", common.MiddlewareAccessTokenAuth)

	v0.POST("/search/subjects", h.Search)

	v0.GET("/subjects/:id", subjectHandler.Get)
	v0.GET("/subjects/:id/image", subjectHandler.GetImage)
	v0.GET("/subjects/:id/persons", subjectHandler.GetRelatedPersons)
	v0.GET("/subjects/:id/subjects", subjectHandler.GetRelatedSubjects)
	v0.GET("/subjects/:id/characters", subjectHandler.GetRelatedCharacters)
	v0.GET("/persons/:id", personHandler.Get)
	v0.GET("/persons/:id/image", personHandler.GetImage)
	v0.GET("/persons/:id/subjects", personHandler.GetRelatedSubjects)
	v0.GET("/persons/:id/characters", personHandler.GetRelatedCharacters)
	v0.GET("/characters/:id", characterHandler.Get)
	v0.GET("/characters/:id/image", characterHandler.GetImage)
	v0.GET("/characters/:id/subjects", characterHandler.GetRelatedSubjects)
	v0.GET("/characters/:id/persons", characterHandler.GetRelatedPersons)
	v0.GET("/episodes/:id", h.GetEpisode)
	v0.GET("/episodes", h.ListEpisode)

	// echo 中间件从前往后运行按顺序
	v0.GET("/me", userHandler.GetCurrent)
	v0.GET("/users/:username", userHandler.Get)
	v0.GET("/users/:username/collections", userHandler.ListSubjectCollection)
	v0.GET("/users/:username/collections/:subject_id", userHandler.GetSubjectCollection)
	v0.GET("/users/-/collections/-/episodes/:episode_id", userHandler.GetEpisodeCollection, accessor.NeedLogin)
	v0.PUT("/users/-/collections/-/episodes/:episode_id", userHandler.PutEpisodeCollection, req.JSON, accessor.NeedLogin)
	v0.GET("/users/-/collections/:subject_id/episodes", userHandler.GetSubjectEpisodeCollection, accessor.NeedLogin)
	v0.PATCH("/users/-/collections/:subject_id", userHandler.PatchSubjectCollection, req.JSON, accessor.NeedLogin)
	v0.PATCH("/users/-/collections/:subject_id/episodes",
		userHandler.PatchEpisodeCollectionBatch, req.JSON, accessor.NeedLogin)

	v0.GET("/users/:username/avatar", userHandler.GetAvatar)

	{
		i := indexHandler
		v0.GET("/indices/:id", i.GetIndex)
		v0.GET("/indices/:id/subjects", i.GetIndexSubjects)
		// indices
		v0.POST("/indices", i.NewIndex, req.JSON, accessor.NeedLogin)
		v0.PUT("/indices/:id", i.UpdateIndex, req.JSON, accessor.NeedLogin)
		// indices subjects
		v0.POST("/indices/:id/subjects", i.AddIndexSubject, req.JSON, accessor.NeedLogin)
		v0.PUT("/indices/:id/subjects/:subject_id", i.UpdateIndexSubject, req.JSON, accessor.NeedLogin)
		v0.DELETE("/indices/:id/subjects/:subject_id", i.RemoveIndexSubject, accessor.NeedLogin)
	}

	v0.GET("/revisions/persons/:id", h.GetPersonRevision)
	v0.GET("/revisions/persons", h.ListPersonRevision)
	v0.GET("/revisions/subjects/:id", h.GetSubjectRevision)
	v0.GET("/revisions/subjects", h.ListSubjectRevision)
	v0.GET("/revisions/characters/:id", h.GetCharacterRevision)
	v0.GET("/revisions/characters", h.ListCharacterRevision)

	v0.GET("/revisions/episodes/:id", h.GetEpisodeRevision)
	v0.GET("/revisions/episodes", h.ListEpisodeRevision)

	var originMiddleware = origin.New(fmt.Sprintf("https://%s", c.WebDomain))
	var refererMiddleware = referer.New(fmt.Sprintf("https://%s/", c.WebDomain))

	var CORSBlockMiddleware []echo.MiddlewareFunc
	if c.WebDomain != "" {
		CORSBlockMiddleware = []echo.MiddlewareFunc{originMiddleware, refererMiddleware}
	}

	// frontend private api
	private := app.Group("/p", append(CORSBlockMiddleware, common.MiddlewareSessionAuth)...)

	// TODO migrate this to bangumi/graphql
	private.GET("/pms/list", pmHandler.List, accessor.NeedLogin)
	private.GET("/pms/related-msgs/:id", pmHandler.ListRelated, accessor.NeedLogin)
	private.GET("/pms/counts", pmHandler.CountTypes, accessor.NeedLogin)
	private.GET("/pms/contacts/recent", pmHandler.ListRecentContact, accessor.NeedLogin)
	private.PATCH("/pms/read", pmHandler.MarkRead)
	private.POST("/pms", pmHandler.Create)
	private.DELETE("/pms", pmHandler.Delete, req.JSON, accessor.NeedLogin)

	private.GET("/notifications/count", notificationHandler.Count, accessor.NeedLogin)

	// default 404 Handler, all router should be added before this router
	app.RouteNotFound("/*", func(c echo.Context) error {
		return c.JSON(http.StatusNotFound, res.Error{
			Title:       "Not Found",
			Description: "This is default response, if you see this response, please check your request",
			Details:     util.Detail(c),
		})
	})
}
