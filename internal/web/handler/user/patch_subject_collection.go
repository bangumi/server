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

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/ctrl"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h User) PatchSubjectCollection(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)

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

	s, err := h.ctrl.GetSubject(c.Context(), u.Auth, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("subject not found")
		}

		h.log.Error("failed to get subject", zap.Error(err), log.SubjectID(subjectID))
		return errgo.Wrap(err, "query.GetSubject")
	}

	if s.TypeID != model.SubjectTypeBook {
		switch {
		case r.VolStatus.Set:
			return res.BadRequest("can't set 'vol_status' on non-book subject")
		case r.EpStatus.Set:
			return res.BadRequest("can't set 'ep_status' on non-book subject")
		}
	}

	collect, err := h.collect.GetSubjectCollection(c.Context(), u.ID, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.BadRequest("subject is not collected")
		}

		return errgo.Wrap(err, "collectionRepo.GetSubjectCollection")
	}

	err = h.ctrl.UpdateCollection(c.Context(), u.Auth, subjectID, ctrl.UpdateCollectionRequest{
		VolStatus: r.VolStatus.Default(collect.VolStatus),
		EpStatus:  r.EpStatus.Default(collect.EpStatus),
		Type:      r.Type.Default(collect.Type),
	})

	return nil
}
