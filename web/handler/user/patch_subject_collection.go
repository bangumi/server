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
	"github.com/trim21/errgo"

	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h User) PatchSubjectCollection(c echo.Context) error {
	subjectID, err := req.ParseID(c.Param("subject_id"))
	if err != nil {
		return err
	}

	var r req.SubjectEpisodeCollectionPatch
	if err = c.Echo().JSONSerializer.Deserialize(c, &r); err != nil {
		return res.JSONError(c, err)
	}

	if err = r.Validate(); err != nil {
		return err
	}

	return h.patchSubjectCollection(c, subjectID, r)
}

func (h User) patchSubjectCollection(
	c echo.Context,
	subjectID model.SubjectID,
	r req.SubjectEpisodeCollectionPatch,
) error {
	u := accessor.GetFromCtx(c)

	s, err := h.subject.Get(c.Request().Context(), subjectID, subject.Filter{NSFW: null.Bool{Set: !u.AllowNSFW()}})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("subject not found")
		}
		return errgo.Wrap(err, "query.GetSubject")
	}

	if s.TypeID != model.SubjectTypeBook {
		if r.VolStatus.Set || r.EpStatus.Set {
			return res.BadRequest("can't set 'vol_status' or 'ep_status' on non-book subject")
		}
	}

	err = h.ctrl.UpdateSubjectCollection(c.Request().Context(), u.Auth, subjectID, ctrl.UpdateCollectionRequest{
		IP:        u.IP,
		UID:       u.ID,
		VolStatus: r.VolStatus,
		EpStatus:  r.EpStatus,
		Type:      r.Type,
		Tags:      r.Tags,
		Comment:   r.Comment,
		Rate:      r.Rate,
		Private:   r.Private,
	})
	if err != nil {
		switch {
		case errors.Is(err, gerr.ErrSubjectNotCollected):
			return res.NotFound("subject not collected")
		case errors.Is(err, gerr.ErrSubjectNotFound):
			return res.NotFound("subject not found")
		}
		return errgo.Wrap(err, "ctrl.UpdateSubjectCollection")
	}

	return c.NoContent(http.StatusNoContent)
}
