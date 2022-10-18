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

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) AddIndexSubject(c *fiber.Ctx) error {
	var reqData req.IndexAddSubject
	if err := json.UnmarshalNoEscape(c.Body(), &reqData); err != nil {
		return res.JSONError(c, err)
	}
	return h.addOrUpdateIndexSubject(c, reqData)
}

func (h Handler) UpdateIndexSubject(c *fiber.Ctx) error {
	var reqData req.IndexSubjectInfo
	if err := json.UnmarshalNoEscape(c.Body(), &reqData); err != nil {
		return res.JSONError(c, err)
	}
	subjectID, err := req.ParseSubjectID(c.Params("subject_id"))
	if err != nil {
		return errgo.Wrap(err, "subject id is invalid")
	}
	return h.addOrUpdateIndexSubject(c, req.IndexAddSubject{
		SubjectID:        subjectID,
		IndexSubjectInfo: &reqData,
	})
}

func (h Handler) addOrUpdateIndexSubject(c *fiber.Ctx, payload req.IndexAddSubject) error {
	indexID, err := req.ParseIndexID(c.Params("id"))
	if err != nil {
		return err
	}
	index, err := h.ensureIndexPermission(c, indexID)
	if err != nil {
		return err
	}
	indexSubject, err := h.i.AddOrUpdateIndexSubject(c.UserContext(),
		index.ID, payload.SubjectID, payload.SortKey, payload.Comment)
	if err != nil {
		if errors.Is(err, domain.ErrSubjectNotFound) {
			return res.NotFound("subject not found")
		}
		return errgo.Wrap(err, "failed to edit subject in the index")
	}
	return c.JSON(indexSubjectToResp(*indexSubject))
}

func (h Handler) RemoveIndexSubject(c *fiber.Ctx) error {
	indexID, err := req.ParseIndexID(c.Params("id"))
	if err != nil {
		return err
	}
	index, err := h.ensureIndexPermission(c, indexID)
	if err != nil {
		return err
	}
	subjectID, err := req.ParseSubjectID(c.Params("subject_id"))
	if err != nil {
		return errgo.Wrap(err, "subject id is invalid")
	}
	if err = h.i.DeleteIndexSubject(c.UserContext(), index.ID, subjectID); err != nil {
		return errgo.Wrap(err, "failed to delete subject from index")
	}
	return nil
}
