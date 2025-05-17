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

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h User) DeleteSubjectCollection(c echo.Context) error {
	subjectID, err := req.ParseID(c.Param("subject_id"))
	if err != nil {
		return err
	}

	return h.deleteSubjectCollection(c, subjectID)
}

func (h User) deleteSubjectCollection(c echo.Context, subjectID model.SubjectID) error {
	u := accessor.GetFromCtx(c)

	err := h.collect.DeleteSubjectCollection(c.Request().Context(), u.ID, subjectID)
	if err != nil {
		switch {
		case errors.Is(err, gerr.ErrSubjectNotCollected):
			return res.JSONError(c, err)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusNoContent)
}
