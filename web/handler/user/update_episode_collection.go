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

package user

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

type ReqEpisodeCollectionBatch struct {
	EpisodeID []model.EpisodeID            `json:"episode_id"`
	Type      collection.EpisodeCollection `json:"type"`
}

func (r ReqEpisodeCollectionBatch) Validate() error {
	if len(r.EpisodeID) == 0 {
		return res.BadRequest("episode_id is required")
	}

	switch r.Type {
	case collection.EpisodeCollectionAll,
		collection.EpisodeCollectionWish,
		collection.EpisodeCollectionDone,
		collection.EpisodeCollectionDropped:
	default:
		return res.BadRequest(fmt.Sprintf("not valid episode collection type %d", r.Type))
	}

	return nil
}

// PatchEpisodeCollectionBatch
//
//	/v0/users/-/collections/:subject_id/episodes"
func (h User) PatchEpisodeCollectionBatch(c echo.Context) error {
	var r ReqEpisodeCollectionBatch
	if err := c.Echo().JSONSerializer.Deserialize(c, &r); err != nil {
		return res.JSONError(c, err)
	}

	if err := r.Validate(); err != nil {
		return err
	}

	subjectID, err := req.ParseID(c.Param("subject_id"))
	if err != nil {
		return err
	}

	u := accessor.GetFromCtx(c)
	err = h.ctrl.UpdateEpisodesCollection(c.Request().Context(), u.Auth, subjectID, r.EpisodeID, r.Type)
	if err != nil {
		switch {
		case errors.Is(err, gerr.ErrSubjectNotCollected):
			return res.BadRequest("you need to add subject to your collection first")
		case errors.Is(err, ctrl.ErrInvalidInput):
			return res.BadRequest(err.Error())
		case errors.Is(err, gerr.ErrNotFound):
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to update episode")
	}

	return c.NoContent(http.StatusNoContent)
}

// PutEpisodeCollection
//
//	/v0/users/-/collections/-/episodes/:episode_id
func (h User) PutEpisodeCollection(c echo.Context) error {
	episodeID, err := req.ParseID(c.Param("episode_id"))
	if err != nil {
		return err
	}

	var r req.UpdateUserEpisodeCollection
	if err = c.Echo().JSONSerializer.Deserialize(c, &r); err != nil {
		return res.JSONError(c, err)
	}

	u := accessor.GetFromCtx(c)
	err = h.ctrl.UpdateEpisodeCollection(c.Request().Context(), u.Auth, episodeID, r.Type)
	if err != nil {
		switch {
		case errors.Is(err, gerr.ErrSubjectNotCollected):
			return res.BadRequest("you need to add subject to your collection first")
		case errors.Is(err, ctrl.ErrInvalidInput):
			return res.BadRequest(err.Error())
		case errors.Is(err, gerr.ErrNotFound):
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to update episode")
	}

	return c.NoContent(http.StatusNoContent)
}
