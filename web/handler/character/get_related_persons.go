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

func (h Character) GetRelatedPersons(c echo.Context) error {
	u := accessor.FromCtx(c)
	id, err := req.ParseCharacterID(c.Param("id"))
	if err != nil {
		return err
	}

	_, err = h.ctrl.GetCharacterNoRedirect(c.Request().Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "failed to get character")
	}

	casts, err := h.person.GetCharacterRelated(c.Request().Context(), id)
	if err != nil {
		return errgo.Wrap(err, "repo.GetCharacterRelated")
	}

	var response = make([]res.CharacterRelatedPerson, len(casts))
	for i, cast := range casts {
		response[i] = res.CharacterRelatedPerson{
			ID:            cast.Person.ID,
			Name:          cast.Person.Name,
			Type:          cast.Person.Type,
			Images:        res.PersonImage(cast.Subject.Image),
			SubjectID:     cast.Subject.ID,
			SubjectName:   cast.Subject.Name,
			SubjectNameCn: cast.Subject.NameCN,
		}
	}

	return c.JSON(http.StatusOK, response)
}
