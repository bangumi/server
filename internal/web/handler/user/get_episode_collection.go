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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

type ResUserEpisodeCollection struct {
	Episode res.Episode             `json:"episode"`
	Type    model.EpisodeCollection `json:"type"`
}

func (h User) GetEpisodeCollection(c *fiber.Ctx) error {
	v := h.GetHTTPAccessor(c)
	episodeID, err := req.ParseEpisodeID(c.Params("episode_id"))
	if err != nil {
		return err
	}

	ec, episode, err := h.app.Query.GetUserEpisodeCollection(c.Context(), v.Auth, episodeID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEpisodeNotFound):
			return res.NotFound("episode not found")
		case errors.Is(err, domain.ErrSubjectNotFound):
			return res.NotFound("subject not exist or has been removed")
		case errors.Is(err, domain.ErrSubjectNotCollected):
			return res.BadRequest("subject is not collected, please add subject to your collection first")
		}

		return h.InternalError(c, err, "failed to get episode collection", log.UserID(v.ID), log.EpisodeID(episodeID))
	}

	return res.JSON(c, ResUserEpisodeCollection{
		Episode: res.ConvertModelEpisode(episode),
		Type:    ec.Type,
	})
}
