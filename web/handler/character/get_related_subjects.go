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

package character

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Character) GetRelatedSubjects(c echo.Context) error {
	u := accessor.FromCtx(c)
	id, err := req.ParseCharacterID(c.Param("id"))
	if err != nil {
		return err
	}

	_, relations, err := h.ctrl.GetCharacterRelatedSubjects(c.Request().Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "repo.GetCharacterRelated")
	}

	var response = make([]res.CharacterRelatedSubject, len(relations))
	for i, relation := range relations {
		subject := relation.Subject
		response[i] = res.CharacterRelatedSubject{
			ID:     subject.ID,
			Name:   subject.Name,
			NameCn: subject.NameCN,
			Staff:  characterStaffString(relation.TypeID),
			Image:  res.SubjectImage(subject.Image).Large,
		}
	}

	return c.JSON(http.StatusOK, response)
}

func characterStaffString(i uint8) string {
	switch i {
	case 1:
		return "主角"
	case 2:
		return "配角"
	case 3:
		return "客串"
	}

	return ""
}
