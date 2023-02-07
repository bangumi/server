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

package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Handler) GetEpisodeRevision(c echo.Context) error {
	id, err := gstr.ParseUint32(c.Param("id"))
	if err != nil || id <= 0 {
		return res.NewError(
			http.StatusBadRequest,
			fmt.Sprintf("bad param id: %s", strconv.Quote(c.Param("id"))),
		)
	}
	r, err := h.r.GetEpisodeRelated(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get episode related revision")
	}

	creatorMap, err := h.u.GetByIDs(c.Request().Context(), []model.UserID{r.CreatorID})
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	return c.JSON(http.StatusOK, convertModelEpisodeRevision(&r, creatorMap))
}

func (h Handler) ListEpisodeRevision(c echo.Context) error {
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	episodeID, err := req.ParseID(c.QueryParam("episode_id"))
	if err != nil {
		return err
	}

	return h.listEpisodeRevision(c, episodeID, page)
}

func (h Handler) listEpisodeRevision(c echo.Context, episodeID model.EpisodeID, page req.PageQuery) error {
	var response = res.Paged{
		Limit:  page.Limit,
		Offset: page.Offset,
	}
	count, err := h.r.CountEpisodeRelated(c.Request().Context(), episodeID)
	if err != nil {
		return errgo.Wrap(err, "revision.CountEpisodeRelated")
	}

	if count == 0 {
		response.Data = []int{}
		return c.JSON(http.StatusOK, response)
	}

	if err = page.Check(count); err != nil {
		return err
	}

	response.Total = count

	revisions, err := h.r.ListEpisodeRelated(c.Request().Context(), episodeID, page.Limit, page.Offset)

	if err != nil {
		return errgo.Wrap(err, "revision.ListEpisodeRelated")
	}

	creatorIDs := make([]model.UserID, 0, len(revisions))
	for _, revision := range revisions {
		creatorIDs = append(creatorIDs, revision.CreatorID)
	}
	creatorMap, err := h.u.GetByIDs(c.Request().Context(), lo.Uniq(creatorIDs))

	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	data := make([]res.EpisodeRevision, len(revisions))
	for i := range revisions {
		data[i] = convertModelEpisodeRevision(&revisions[i], creatorMap)
	}
	response.Data = data
	return c.JSON(http.StatusOK, response)
}
