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

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

type ResUserEpisodeCollection struct {
	Episode res.Episode             `json:"episode"`
	Type    model.EpisodeCollection `json:"type"`
}

func (h User) GetEpisodeCollection(c echo.Context) error {
	v := accessor.GetFromCtx(c)
	episodeID, err := req.ParseEpisodeID(c.Param("episode_id"))
	if err != nil {
		return err
	}

	e, err := h.episode.Get(c.Request().Context(), episodeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		h.log.Error("failed to get episode", episodeID.Zap())
		return errgo.Wrap(err, "query.GetEpisode")
	}

	m, err := h.collect.GetSubjectEpisodesCollection(c.Request().Context(), v.ID, e.SubjectID)
	if err != nil {
		return errgo.Wrap(err, "collectionRepo.GetSubjectEpisodesCollection")
	}

	return c.JSON(http.StatusOK, ResUserEpisodeCollection{
		Episode: res.ConvertModelEpisode(e),
		Type:    m[episodeID].Type,
	})
}

// GetSubjectEpisodeCollection return episodes with user's collection info.
func (h User) GetSubjectEpisodeCollection(c echo.Context) error {
	v := accessor.GetFromCtx(c)
	subjectID, err := req.ParseSubjectID(c.Param("subject_id"))
	if err != nil {
		return err
	}

	page, err := req.GetPageQuery(c, req.EpisodeDefaultLimit, req.EpisodeMaxLimit)
	if err != nil {
		return err
	}

	episodeType, err := req.ParseEpTypeOptional(c.QueryParam("episode_type"))
	if err != nil {
		return err
	}

	_, err = h.subject.Get(c.Request().Context(), subjectID, subject.Filter{NSFW: null.Bool{Set: !v.AllowNSFW()}})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		h.log.Error("failed to fetch subject", zap.Error(err), subjectID.Zap(), v.Log())
		return errgo.Wrap(err, "query.GetSubject")
	}

	ec, err := h.collect.GetSubjectEpisodesCollection(c.Request().Context(), v.ID, subjectID)
	if err != nil {
		h.log.Error("unexpected error to fetch user subject collections",
			zap.Error(err), v.ID.Zap(), subjectID.Zap())
		return errgo.Wrap(err, "collectionRepo.GetSubjectEpisodesCollection")
	}

	count, err := h.episode.Count(c.Request().Context(), subjectID, episode.Filter{Type: null.NewFromPtr(episodeType)})
	if err != nil {
		return errgo.Wrap(err, "count episodes")
	}

	episodes, err := h.episode.List(c.Request().Context(), subjectID,
		episode.Filter{Type: null.NewFromPtr(episodeType)}, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "list episodes")
	}

	var data []ResUserEpisodeCollection

	for _, episode := range episodes {
		data = append(data, ResUserEpisodeCollection{
			Episode: res.ConvertModelEpisode(episode),
			Type:    ec[episode.ID].Type,
		})
	}

	return c.JSON(http.StatusOK, res.PagedG[ResUserEpisodeCollection]{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}
