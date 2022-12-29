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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Subject) GetRelatedSubjects(c *fiber.Ctx) error {
	id, err := req.ParseSubjectID(c.Params("id"))
	if err != nil {
		return err
	}

	u := h.GetHTTPAccessor(c)

	relations, err := h.ctrl.GetSubjectRelatedSubjects(c.UserContext(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "repo")
	}

	var response = make([]res.SubjectRelatedSubject, len(relations))
	for i, relation := range relations {
		response[i] = res.SubjectRelatedSubject{
			Images:    res.SubjectImage(relation.Destination.Image),
			Name:      relation.Destination.Name,
			NameCn:    relation.Destination.NameCN,
			Relation:  readableRelation(relation.Destination.TypeID, relation.TypeID),
			Type:      relation.Destination.TypeID,
			SubjectID: relation.Destination.ID,
		}
	}

	return c.JSON(response)
}
