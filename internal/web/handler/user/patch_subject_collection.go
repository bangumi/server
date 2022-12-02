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

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/ctrl"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h User) PatchSubjectCollection(c *fiber.Ctx) error {
	subjectID, err := req.ParseSubjectID(c.Params("subject_id"))
	if err != nil {
		return err
	}

	var r req.SubjectEpisodeCollectionPatch
	if err = sonic.Unmarshal(c.Body(), &r); err != nil {
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

	s, err := h.ctrl.GetSubject(c.UserContext(), u.Auth, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("subject not found")
		}

		h.log.Error("failed to get subject", zap.Error(err), subjectID.Zap())
		return errgo.Wrap(err, "query.GetSubject")
	}

	if s.TypeID != model.SubjectTypeBook {
		if r.VolStatus.Set || r.EpStatus.Set {
			return res.BadRequest("can't set 'vol_status' or 'ep_status' on non-book subject")
		}
	}

	err = h.ctrl.UpdateCollection(c.UserContext(), u.Auth, subjectID, ctrl.UpdateCollectionRequest{
		IP:        u.IP.String(),
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
		case errors.Is(err, domain.ErrSubjectNotCollected):
			return res.NotFound("subject not collected")
		case errors.Is(err, domain.ErrSubjectNotFound):
			return res.NotFound("subject not found")
		}
		return errgo.Wrap(err, "ctrl.UpdateCollection")
	}

	c.Status(http.StatusNoContent)
	return nil
}
