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
	"time"

	"github.com/labstack/echo/v5"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/pkg/vars"
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

	metaTags, err := h.tag.Get(c.Request().Context(), s.ID, s.TypeID)
	if err != nil {
		return err
	}

	if !s.NSFW {
		res.SetCacheControl(c, res.CacheControlParams{Public: true, MaxAge: time.Hour})
	}

	return c.JSON(http.StatusOK, res.ToSubjectV0(s, totalEpisode, metaTags))
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

	res.SetCacheControl(c, res.CacheControlParams{Public: true, MaxAge: time.Hour})

	if l == "" {
		return c.Redirect(http.StatusFound, res.DefaultImageURL)
	}

	return c.Redirect(http.StatusFound, l)
}

func readableRelation(destSubjectType model.SubjectType, relation uint16) string {
	var r, ok = vars.RelationMap[destSubjectType][relation]
	if !ok || relation == 1 {
		return model.SubjectTypeString(destSubjectType)
	}

	return r.String(relation)
}
