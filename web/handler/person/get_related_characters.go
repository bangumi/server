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

package person

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Person) GetRelatedCharacters(c echo.Context) error {
	id, err := req.ParsePersonID(c.Param("id"))
	if err != nil {
		return err
	}

	r, err := h.ctrl.GetPerson(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get person")
	}

	if r.Redirect != 0 {
		return res.ErrNotFound
	}

	relations, err := h.ctrl.GetPersonRelatedCharacters(c.Request().Context(), id)
	if err != nil {
		return errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	var response = make([]res.PersonRelatedCharacter, len(relations))
	for i, rel := range relations {
		response[i] = res.PersonRelatedCharacter{
			ID:            rel.Character.ID,
			Name:          rel.Character.Name,
			Type:          rel.Character.Type,
			Images:        res.PersonImage(rel.Subject.Image),
			SubjectID:     rel.Subject.ID,
			SubjectName:   rel.Subject.Name,
			SubjectNameCn: rel.Subject.NameCN,
		}
	}

	return c.JSON(http.StatusOK, response)
}
