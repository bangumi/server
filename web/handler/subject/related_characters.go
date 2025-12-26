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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Subject) GetRelatedCharacters(c echo.Context) error {
	u := accessor.GetFromCtx(c)
	subjectID, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	_, relations, err := h.getSubjectRelatedCharacters(c.Request().Context(), u.Auth, subjectID)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "CharacterRepo.GetSubjectRelated")
	}

	var actors map[model.CharacterID][]model.Person
	if len(relations) != 0 {
		var characterIDs = slice.Map(relations,
			func(item model.SubjectCharacterRelation) model.CharacterID { return item.Character.ID })
		actors, err = h.getActors(c.Request().Context(), subjectID, characterIDs...)
		if err != nil {
			return errgo.Wrap(err, "query.getActors")
		}
	}

	var response = make([]res.SubjectRelatedCharacter, len(relations))
	for i, rel := range relations {
		response[i] = res.SubjectRelatedCharacter{
			Images:   res.PersonImage(rel.Character.Image),
			Name:     rel.Character.Name,
			Summary:  rel.Character.Summary,
			Relation: res.CharacterStaffString(rel.TypeID),
			Actors:   toActors(actors[rel.Character.ID]),
			Type:     rel.Character.Type,
			ID:       rel.Character.ID,
		}
	}
	res.SetCacheControl(c, res.CacheControlParams{Public: true, MaxAge: time.Hour})

	return c.JSON(http.StatusOK, response)
}

func (h Subject) getActors(
	ctx context.Context,
	subjectID model.SubjectID,
	characterIDs ...model.CharacterID,
) (map[model.CharacterID][]model.Person, error) {
	actors, err := h.subject.GetActors(ctx, subjectID, characterIDs)
	if err != nil {
		return nil, errgo.Wrap(err, "subjectRepo.getActors")
	}

	vs := lo.Uniq(lo.Flatten(lo.Values(actors)))

	persons, err := h.personRepo.GetByIDs(ctx, vs)
	if err != nil {
		return nil, errgo.Wrap(err, "failed to get persons")
	}

	var result = make(map[model.CharacterID][]model.Person, len(actors))

	for characterID, ids := range actors {
		result[characterID] = slice.Map(ids, func(item model.PersonID) model.Person {
			return persons[item]
		})
	}

	return result, nil
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
		if errors.Is(err, gerr.ErrNotFound) {
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
