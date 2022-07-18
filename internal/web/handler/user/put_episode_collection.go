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
	"net/http"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h User) PutEpisodeCollection(c *fiber.Ctx) error {
	v := h.GetHTTPAccessor(c)
	if !v.Login {
		return res.Unauthorized(res.DefaultUnauthorizedMessage)
	}

	episodeID, err := req.ParseEpisodeID(c.Params("episode_id"))
	if err != nil {
		return err
	}

	var input req.PutEpisodeCollection
	if err = json.UnmarshalNoEscape(c.Body(), &input); err != nil {
		return res.JSONError(c, err)
	}

	if errs := h.Common.V.Struct(input); errs != nil {
		return h.ValidationError(c, errs)
	}

	// now call app command
	if err = h.app.Command.UpdateEpisodeCollection(c.Context(), v.ID, episodeID, input.Type); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return res.FromError(c, err, http.StatusBadRequest, "failed to update episode collection")
		case errors.Is(err, domain.ErrEpisodeNotFound):
			return res.NotFound("episode not found")
		case errors.Is(err, domain.ErrSubjectNotFound):
			return res.NotFound("subject not exist or has been removed")
		case errors.Is(err, domain.ErrSubjectNotCollected):
			return res.BadRequest("subject is not collected, please add subject to your collection first")
		}

		return h.InternalError(c, err, "failed to update episode collection",
			log.UserID(v.ID), log.EpisodeID(episodeID))
	}

	c.Status(http.StatusNoContent)
	return nil
}
