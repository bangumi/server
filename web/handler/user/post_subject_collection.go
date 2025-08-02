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

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h User) PostSubjectCollection(c echo.Context) error {
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

	// 与 PatchSubjectCollection 一致
	// 但允许创建，如果不存在
	if err := h.updateOrCreateSubjectCollection(c, subjectID, r, true); err != nil {
		switch {
		case errors.Is(err, gerr.ErrSubjectNotCollected):
			return res.NotFound("subject not collected")
		case errors.Is(err, gerr.ErrSubjectNotFound):
			return res.NotFound("subject not found")
		case errors.Is(err, gerr.ErrBanned):
			return res.Forbidden(err.Error())
		}
		return errgo.Wrap(err, "ctrl.UpdateSubjectCollection")
	}
	return c.NoContent(http.StatusAccepted)
}
