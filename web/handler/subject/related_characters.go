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
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Subject) GetRelatedCharacters(c echo.Context) error {
	u := accessor.FromCtx(c)
	subjectID, err := req.ParseSubjectID(c.Param("id"))
	if err != nil {
		return err
	}

	_, relations, err := h.getSubjectRelatedCharacters(c.Request().Context(), u.Auth, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "CharacterRepo.GetSubjectRelated")
	}

	var actors map[model.CharacterID][]model.Person
	if len(relations) != 0 {
		var characterIDs = slice.Map(relations,
			func(item model.SubjectCharacterRelation) model.CharacterID { return item.Character.ID })
		actors, err = h.ctrl.GetActors(c.Request().Context(), subjectID, characterIDs...)
		if err != nil {
			return errgo.Wrap(err, "query.GetActors")
		}
	}

	var response = make([]res.SubjectRelatedCharacter, len(relations))
	for i, rel := range relations {
		response[i] = res.SubjectRelatedCharacter{
			Images:   res.PersonImage(rel.Character.Image),
			Name:     rel.Character.Name,
			Relation: characterStaffString(rel.TypeID),
			Actors:   toActors(actors[rel.Character.ID]),
			Type:     rel.Character.Type,
			ID:       rel.Character.ID,
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (h Subject) getSubjectRelatedCharacters(
	ctx context.Context,
	user auth.Auth,
	subjectID model.SubjectID,
) (model.Subject, []model.SubjectCharacterRelation, error) {
	s, err := h.subject.Get(ctx, subjectID, subject.Filter{NSFW: null.Bool{
		Value: false, Set: !user.AllowNSFW(),
	}})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return s, nil, res.ErrNotFound
		}
		return s, nil, errgo.Wrap(err, "failed to get subject")
	}

	relations, err := h.c.GetSubjectRelated(ctx, subjectID)
	if err != nil {
		return s, nil, errgo.Wrap(err, "CharacterRepo.GetSubjectRelated")
	}

	var characterIDs = slice.Map(relations, func(item domain.SubjectCharacterRelation) model.CharacterID {
		return item.CharacterID
	})

	characters, err := h.c.GetByIDs(ctx, characterIDs)
	if err != nil {
		return s, nil, errgo.Wrap(err, "CharacterRepo.GetByIDs")
	}

	var results = make([]model.SubjectCharacterRelation, len(relations))
	for i, rel := range relations {
		results[i] = model.SubjectCharacterRelation{
			Character: characters[rel.CharacterID],
			TypeID:    rel.TypeID,
		}
	}

	return s, results, nil
}

func toActors(persons []model.Person) []res.Actor {
	// should pre-alloc a big slice and split it into sub slice.
	var actors = make([]res.Actor, len(persons))
	for j, actor := range persons {
		actors[j] = res.Actor{
			Images:       res.PersonImage(actor.Image),
			Name:         actor.Name,
			ShortSummary: actor.Summary,
			Career:       actor.Careers(),
			ID:           actor.ID,
			Type:         actor.Type,
			Locked:       actor.Locked,
		}
	}

	return actors
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
