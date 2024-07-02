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

package subject

import (
	"errors"
	"github.com/bangumi/server/pkg/wiki"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/pkg/vars"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Subject) GetRelatedPersons(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil || id == 0 {
		return err
	}

	r, err := h.subject.Get(c.Request().Context(), id, subject.Filter{
		NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()},
	})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "failed to get subject")
	}

	relations, err := h.person.GetSubjectRelated(c.Request().Context(), id)
	if err != nil {
		return errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	var response = make([]res.SubjectRelatedPerson, len(relations))
	for i, rel := range relations {
		response[i] = res.SubjectRelatedPerson{
			Images:   res.PersonImage(rel.Person.Image),
			Name:     rel.Person.Name,
			NameCN:   wiki.ParseNameCN(rel.Person.Infobox),
			Relation: vars.StaffMap[r.TypeID][rel.TypeID].String(),
			Career:   rel.Person.Careers(),
			Type:     rel.Person.Type,
			ID:       rel.Person.ID,
		}
	}

	return c.JSON(http.StatusOK, response)
}
