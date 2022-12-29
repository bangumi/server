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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h User) GetSubjectCollection(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}

	subjectID, err := req.ParseSubjectID(c.Params("subject_id"))
	if err != nil {
		return err
	}

	return h.getSubjectCollection(c, username, subjectID)
}

func (h User) getSubjectCollection(c *fiber.Ctx, username string, subjectID model.SubjectID) error {
	const notFoundMessage = "subject is not collected by user"
	v := h.GetHTTPAccessor(c)

	s, err := h.ctrl.GetSubject(c.UserContext(), v.Auth, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to subject info")
	}

	u, err := h.user.GetByName(c.UserContext(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("user doesn't exist or has been removed")
		}

		return errgo.Wrap(err, "failed to get user by name")
	}

	var showPrivate = u.ID == v.ID

	collection, err := h.collect.GetSubjectCollection(c.UserContext(), u.ID, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound(notFoundMessage)
		}

		return errgo.Wrap(err, "failed to get user's subject collection")
	}

	if !showPrivate && collection.Private {
		return res.NotFound(notFoundMessage)
	}

	return res.JSON(c, res.ConvertModelSubjectCollection(collection, res.ToSlimSubjectV0(s)))
}
