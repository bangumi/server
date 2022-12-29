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
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
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

	e, err := h.ctrl.GetEpisode(c.UserContext(), episodeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		h.log.Error("failed to get episode", episodeID.Zap())
		return errgo.Wrap(err, "query.GetEpisode")
	}

	m, err := h.collect.GetSubjectEpisodesCollection(c.UserContext(), v.ID, e.SubjectID)
	if err != nil {
		return errgo.Wrap(err, "collectionRepo.GetSubjectEpisodesCollection")
	}

	return res.JSON(c, ResUserEpisodeCollection{
		Episode: res.ConvertModelEpisode(e),
		Type:    m[episodeID].Type,
	})
}

// GetSubjectEpisodeCollection return episodes with user's collection info.
func (h User) GetSubjectEpisodeCollection(c *fiber.Ctx) error {
	v := h.GetHTTPAccessor(c)
	subjectID, err := req.ParseSubjectID(c.Params("subject_id"))
	if err != nil {
		return err
	}

	page, err := req.GetPageQuery(c, req.EpisodeDefaultLimit, req.EpisodeMaxLimit)
	if err != nil {
		return err
	}

	episodeType, err := req.ParseEpTypeOptional(c.Query("episode_type"))
	if err != nil {
		return err
	}

	_, err = h.ctrl.GetSubject(c.UserContext(), v.Auth, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		h.log.Error("failed to fetch subject", zap.Error(err), subjectID.Zap(), v.Log())
		return errgo.Wrap(err, "query.GetSubject")
	}

	ec, err := h.collect.GetSubjectEpisodesCollection(c.UserContext(), v.ID, subjectID)
	if err != nil {
		h.log.Error("unexpected error to fetch user subject collections",
			zap.Error(err), v.ID.Zap(), subjectID.Zap())
		return errgo.Wrap(err, "collectionRepo.GetSubjectEpisodesCollection")
	}

	episodes, count, err := h.ctrl.ListEpisode(c.UserContext(), subjectID, episodeType, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "query.ListEpisode")
	}

	var data []ResUserEpisodeCollection

	for _, episode := range episodes {
		data = append(data, ResUserEpisodeCollection{
			Episode: res.ConvertModelEpisode(episode),
			Type:    ec[episode.ID].Type,
		})
	}

	return res.JSON(c, res.PagedG[ResUserEpisodeCollection]{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}
