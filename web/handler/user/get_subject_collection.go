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

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h User) GetSubjectCollection(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}

	subjectID, err := req.ParseID(c.Param("subject_id"))
	if err != nil {
		return err
	}

	return h.getSubjectCollection(c, username, subjectID)
}

func (h User) getSubjectCollection(c echo.Context, username string, subjectID model.SubjectID) error {
	const notFoundMessage = "subject is not collected by user"
	v := accessor.GetFromCtx(c)

	s, err := h.subject.Get(c.Request().Context(), subjectID, subject.Filter{NSFW: null.Bool{Set: !v.AllowNSFW()}})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to subject info")
	}

	u, err := h.user.GetByName(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("user doesn't exist or has been removed")
		}

		return errgo.Wrap(err, "failed to get user by name")
	}

	var showPrivate = u.ID == v.ID

	collection, err := h.collect.GetSubjectCollection(c.Request().Context(), u.ID, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound(notFoundMessage)
		}

		return errgo.Wrap(err, "failed to get user's subject collection")
	}

	if !showPrivate && collection.Private {
		return res.NotFound(notFoundMessage)
	}

	return c.JSON(http.StatusOK, res.ConvertModelSubjectCollection(collection, res.ToSlimSubjectV0(s)))
}
