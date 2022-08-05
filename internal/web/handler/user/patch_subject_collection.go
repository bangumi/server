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

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/ctrl"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h User) PatchSubjectCollection(c *fiber.Ctx) error {
	subjectID, err := req.ParseSubjectID(c.Params("subject_id"))
	if err != nil {
		return err
	}

	var r req.SubjectEpisodeCollectionPatch
	if err = json.Unmarshal(c.Body(), &r); err != nil {
		return res.JSONError(c, err)
	}

	if err = r.Validate(); err != nil {
		return err
	}

	return h.patchSubjectCollection(c, subjectID, r)
}

func (h User) patchSubjectCollection(
	c *fiber.Ctx,
	subjectID model.SubjectID,
	r req.SubjectEpisodeCollectionPatch,
) error {
	u := h.GetHTTPAccessor(c)

	s, err := h.ctrl.GetSubject(c.Context(), u.Auth, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("subject not found")
		}

		h.log.Error("failed to get subject", zap.Error(err), log.SubjectID(subjectID))
		return errgo.Wrap(err, "query.GetSubject")
	}

	if s.TypeID != model.SubjectTypeBook {
		if r.VolStatus.Set || r.EpStatus.Set {
			return res.BadRequest("can't set 'vol_status' or 'ep_status' on non-book subject")
		}
	}

	collect, err := h.collect.GetSubjectCollection(c.Context(), u.ID, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.BadRequest("subject is not collected")
		}
		return errgo.Wrap(err, "collectionRepo.GetSubjectCollection")
	}

	ctrlReq := ctrl.UpdateCollectionRequest{
		IP: u.IP.String(),

		VolStatus: r.VolStatus,
		EpStatus:  r.EpStatus,
		Type:      null.New(model.SubjectCollection(r.Type.Default(uint8(collect.Type)))),
		Tags:      collect.Tags,
		Comment:   r.Comment,
		Rate:      r.Rate,
	}

	if r.Tags != nil {
		ctrlReq.Tags = r.Tags
	}

	err = h.ctrl.UpdateCollection(c.Context(), u.Auth, subjectID, ctrlReq)
	if err != nil {
		return errgo.Wrap(err, "ctrl.UpdateCollection")
	}

	c.Status(http.StatusNoContent)
	return nil
}
