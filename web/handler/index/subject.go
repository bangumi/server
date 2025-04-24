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

package index

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Handler) AddIndexSubject(c echo.Context) error {
	var reqData req.IndexAddSubject
	if err := c.Echo().JSONSerializer.Deserialize(c, &reqData); err != nil {
		return res.JSONError(c, err)
	}

	return h.addOrUpdateIndexSubject(c, reqData)
}

func (h Handler) UpdateIndexSubject(c echo.Context) error {
	var reqData req.IndexSubjectInfo
	if err := c.Echo().JSONSerializer.Deserialize(c, &reqData); err != nil {
		return res.JSONError(c, err)
	}
	subjectID, err := req.ParseID(c.Param("subject_id"))
	if err != nil {
		return errgo.Wrap(err, "subject id is invalid")
	}
	return h.addOrUpdateIndexSubject(c, req.IndexAddSubject{
		SubjectID:        subjectID,
		IndexSubjectInfo: reqData,
	})
}

func (h Handler) addOrUpdateIndexSubject(c echo.Context, payload req.IndexAddSubject) error {
	if err := h.ensureValidStrings(payload.Comment); err != nil {
		return err
	}
	indexID, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	index, err := h.ensureIndexPermission(c, indexID)
	if err != nil {
		return err
	}
	indexSubject, err := h.i.AddOrUpdateIndexSubject(c.Request().Context(),
		index.ID, payload.SubjectID, payload.SortKey, payload.Comment)
	if err != nil {
		if errors.Is(err, gerr.ErrSubjectNotFound) {
			return res.NotFound("subject not found")
		}
		return errgo.Wrap(err, "failed to edit subject in the index")
	}
	return c.JSON(http.StatusOK, indexSubjectToResp(*indexSubject))
}

func (h Handler) RemoveIndexSubject(c echo.Context) error {
	indexID, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	index, err := h.ensureIndexPermission(c, indexID)
	if err != nil {
		return err
	}
	subjectID, err := req.ParseID(c.Param("subject_id"))
	if err != nil {
		return errgo.Wrap(err, "subject id is invalid")
	}
	if err = h.i.DeleteIndexSubject(c.Request().Context(), index.ID, subjectID); err != nil {
		return errgo.Wrap(err, "failed to delete subject from index")
	}
	return nil
}
