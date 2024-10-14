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

package subject

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/compat"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/tag"
	"github.com/bangumi/server/pkg/vars"
	"github.com/bangumi/server/pkg/wiki"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Subject) Get(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	s, err := h.subject.Get(c.Request().Context(), id, subject.Filter{
		NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()},
	})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get subject")
	}

	if s.Redirect != 0 {
		return c.Redirect(http.StatusFound, fmt.Sprintf("/v0/subjects/%d", s.Redirect))
	}

	totalEpisode, err := h.episode.Count(c.Request().Context(), id, episode.Filter{})
	if err != nil {
		return errgo.Wrap(err, "episode.Count")
	}

	metaTags, err := h.tag.Get(c.Request().Context(), s.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, convertModelSubject(s, totalEpisode, metaTags))
}

func platformString(s model.Subject) *string {
	platform, ok := vars.PlatformMap[s.TypeID][s.PlatformID]
	if !ok && s.TypeID != 0 {
		logger.Warn("unknown platform",
			zap.Uint32("subject", s.ID),
			zap.Uint8("type", s.TypeID),
			zap.Uint16("platform", s.PlatformID),
		)

		return nil
	}

	v := platform.String()

	return &v
}

func (h Subject) GetImage(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil || id == 0 {
		return err
	}

	r, err := h.subject.Get(c.Request().Context(), id, subject.Filter{NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()}})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "failed to get subject")
	}

	l, ok := res.SubjectImage(r.Image).Select(c.QueryParam("type"))
	if !ok {
		return res.BadRequest("bad image type: " + c.QueryParam("type"))
	}

	if l == "" {
		return c.Redirect(http.StatusFound, res.DefaultImageURL)
	}

	return c.Redirect(http.StatusFound, l)
}

func convertModelSubject(s model.Subject, totalEpisode int64, metaTags []tag.Tag) res.SubjectV0 {
	return res.SubjectV0{
		TotalEpisodes: totalEpisode,
		ID:            s.ID,
		Image:         res.SubjectImage(s.Image),
		Summary:       s.Summary,
		Name:          s.Name,
		Platform:      platformString(s),
		NameCN:        s.NameCN,
		Date:          null.NilString(s.Date),
		Infobox:       compat.V0Wiki(wiki.ParseOmitError(s.Infobox).NonZero()),
		Volumes:       s.Volumes,
		Redirect:      s.Redirect,
		Eps:           s.Eps,
		MetaTags: lo.Map(metaTags, func(item tag.Tag, index int) string {
			return item.Name
		}),
		Tags: slice.Map(s.Tags, func(tag model.Tag) res.SubjectTag {
			return res.SubjectTag{
				Name:  tag.Name,
				Count: tag.Count,
			}
		}),
		Collection: res.SubjectCollectionStat{
			OnHold:  s.OnHold,
			Wish:    s.Wish,
			Dropped: s.Dropped,
			Collect: s.Collect,
			Doing:   s.Doing,
		},
		TypeID: s.TypeID,
		Series: s.Series,
		Locked: s.Locked(),
		NSFW:   s.NSFW,
		Rating: res.Rating{
			Rank:  s.Rating.Rank,
			Total: s.Rating.Total,
			Count: res.Count{
				Field1:  s.Rating.Count.Field1,
				Field2:  s.Rating.Count.Field2,
				Field3:  s.Rating.Count.Field3,
				Field4:  s.Rating.Count.Field4,
				Field5:  s.Rating.Count.Field5,
				Field6:  s.Rating.Count.Field6,
				Field7:  s.Rating.Count.Field7,
				Field8:  s.Rating.Count.Field8,
				Field9:  s.Rating.Count.Field9,
				Field10: s.Rating.Count.Field10,
			},
			Score: s.Rating.Score,
		},
	}
}

func readableRelation(destSubjectType model.SubjectType, relation uint16) string {
	var r, ok = vars.RelationMap[destSubjectType][relation]
	if !ok || relation == 1 {
		return model.SubjectTypeString(destSubjectType)
	}

	return r.String(relation)
}
