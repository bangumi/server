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

	"github.com/bytedance/sonic/decoder"
	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

type ReqEpisodeCollectionBatch struct {
	EpisodeID []model.EpisodeID       `json:"episode_id"`
	Type      model.EpisodeCollection `json:"type"`
}

func (r ReqEpisodeCollectionBatch) Validate() error {
	if len(r.EpisodeID) == 0 {
		return res.BadRequest("episode_id is required")
	}

	switch r.Type {
	case model.EpisodeCollectionAll,
		model.EpisodeCollectionWish,
		model.EpisodeCollectionDone,
		model.EpisodeCollectionDropped:
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
	if err := decoder.NewStreamDecoder(c.Request().Body).Decode(&r); err != nil {
		return res.JSONError(c, err)
	}

	if err := r.Validate(); err != nil {
		return err
	}

	subjectID, err := req.ParseSubjectID(c.Param("subject_id"))
	if err != nil {
		return err
	}

	u := h.GetHTTPAccessor(c)
	err = h.ctrl.UpdateEpisodesCollection(c.Request().Context(), u.Auth, subjectID, r.EpisodeID, r.Type)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSubjectNotCollected):
			return res.BadRequest("you need to add subject to your collection first")
		case errors.Is(err, ctrl.ErrInvalidInput):
			return res.BadRequest(err.Error())
		case errors.Is(err, domain.ErrNotFound):
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
	episodeID, err := req.ParseEpisodeID(c.Param("episode_id"))
	if err != nil {
		return err
	}

	var r req.UpdateUserEpisodeCollection
	if err = decoder.NewStreamDecoder(c.Request().Body).Decode(&r); err != nil {
		return res.JSONError(c, err)
	}

	u := h.GetHTTPAccessor(c)
	err = h.ctrl.UpdateEpisodeCollection(c.Request().Context(), u.Auth, episodeID, r.Type)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSubjectNotCollected):
			return res.BadRequest("you need to add subject to your collection first")
		case errors.Is(err, ctrl.ErrInvalidInput):
			return res.BadRequest(err.Error())
		case errors.Is(err, domain.ErrNotFound):
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to update episode")
	}

	return c.NoContent(http.StatusNoContent)
}
