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

package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/compat"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/pkg/vars"
	"github.com/bangumi/server/pkg/wiki"
)

func (h Handler) GetSubject(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)

	id, err := req.ParseSubjectID(c.Params("id"))
	if err != nil {
		return err
	}

	s, err := h.app.Query.GetSubject(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get subject", log.SubjectID(id))
	}

	if s.Redirect != 0 {
		return c.Redirect("/v0/subjects/" + strconv.FormatUint(uint64(s.Redirect), 10))
	}

	totalEpisode, err := h.app.Query.CountEpisode(c.Context(), id, nil)
	if err != nil {
		return h.InternalError(c, err, "failed to count episodes of subject", log.SubjectID(id))
	}

	return res.JSON(c, convertModelSubject(s, totalEpisode))
}

func platformString(s model.Subject) *string {
	platform, ok := vars.PlatformMap[s.TypeID][s.PlatformID]
	if !ok && s.TypeID != 0 {
		logger.Warn("unknown platform",
			log.SubjectID(s.ID),
			zap.Uint8("type", s.TypeID),
			zap.Uint16("platform", s.PlatformID),
		)

		return nil
	}

	v := platform.String()

	return &v
}

func (h Handler) GetSubjectImage(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)

	id, err := req.ParseSubjectID(c.Params("id"))
	if err != nil || id == 0 {
		return err
	}

	r, err := h.app.Query.GetSubject(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to get subject", log.SubjectID(id))
	}

	l, ok := res.SubjectImage(r.Image).Select(c.Query("type"))
	if !ok {
		return res.BadRequest("bad image type: " + c.Query("type"))
	}

	if l == "" {
		return c.Redirect(res.DefaultImageURL)
	}

	return c.Redirect(l)
}

func getExpectSubjectID(c *fiber.Ctx, topic model.Topic) (model.SubjectID, error) {
	subjectID, err := req.ParseSubjectID(c.Params("id"))
	if err != nil || subjectID == 0 {
		subjectID = model.SubjectID(topic.ObjectID)
	} else if subjectID != model.SubjectID(topic.ObjectID) {
		return model.SubjectID(0), res.ErrNotFound
	}
	return subjectID, nil
}

func (h Handler) GetSubjectRelatedPersons(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)

	id, err := req.ParseSubjectID(c.Params("id"))
	if err != nil || id == 0 {
		return err
	}

	r, err := h.app.Query.GetSubjectNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to get subject", log.SubjectID(id))
	}

	relations, err := h.p.GetSubjectRelated(c.Context(), id)
	if err != nil {
		return errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	var response = make([]res.SubjectRelatedPerson, len(relations))
	for i, rel := range relations {
		response[i] = res.SubjectRelatedPerson{
			Images:   res.PersonImage(rel.Person.Image),
			Name:     rel.Person.Name,
			Relation: vars.StaffMap[r.TypeID][rel.TypeID].String(),
			Career:   rel.Person.Careers(),
			Type:     rel.Person.Type,
			ID:       rel.Person.ID,
		}
	}

	return c.JSON(response)
}

func convertModelSubject(s model.Subject, totalEpisode int64) res.SubjectV0 {
	return res.SubjectV0{
		TotalEpisodes: totalEpisode,
		ID:            s.ID,
		Image:         res.SubjectImage(s.Image),
		Summary:       s.Summary,
		Name:          s.Name,
		Platform:      platformString(s),
		NameCN:        s.NameCN,
		Date:          nilString(s.Date),
		Infobox:       compat.V0Wiki(wiki.ParseOmitError(s.Infobox).NonZero()),
		Volumes:       s.Volumes,
		Redirect:      s.Redirect,
		Eps:           s.Eps,
		Tags: slice.Map(s.Tags, func(tag model.Tag) res.SubjectTag {
			return res.SubjectTag{
				Name:  tag.Name,
				Count: tag.Count,
			}
		}),
		Collection: res.SubjectCollectionStat{
			OnHold:  s.OnHold,
			Wish:    s.Wish,
			Dropped: s.Dropped,
			Collect: s.Collect,
			Doing:   s.Doing,
		},
		TypeID: s.TypeID,
		Locked: s.Locked(),
		NSFW:   s.NSFW,
		Rating: res.Rating{
			Rank:  s.Rating.Rank,
			Total: s.Rating.Total,
			Count: res.Count{
				Field1:  s.Rating.Count.Field1,
				Field2:  s.Rating.Count.Field2,
				Field3:  s.Rating.Count.Field3,
				Field4:  s.Rating.Count.Field4,
				Field5:  s.Rating.Count.Field5,
				Field6:  s.Rating.Count.Field6,
				Field7:  s.Rating.Count.Field7,
				Field8:  s.Rating.Count.Field8,
				Field9:  s.Rating.Count.Field9,
				Field10: s.Rating.Count.Field10,
			},
			Score: s.Rating.Score,
		},
	}
}

func (h Handler) GetSubjectRelatedSubjects(c *fiber.Ctx) error {
	id, err := req.ParseSubjectID(c.Params("id"))
	if err != nil {
		return err
	}

	u := h.getHTTPAccessor(c)

	_, relations, err := h.app.Query.GetSubjectRelatedSubjects(c.Context(), u.Auth, id)
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

func readableRelation(destSubjectType model.SubjectType, relation uint16) string {
	var r, ok = vars.RelationMap[destSubjectType][relation]
	if !ok || relation == 1 {
		return model.SubjectTypeString(destSubjectType)
	}

	return r.String()
}

func (h Handler) GetSubjectRelatedCharacters(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)
	id, err := req.ParseSubjectID(c.Params("id"))
	if err != nil {
		return err
	}

	_, err = h.app.Query.GetSubjectNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to get subject", log.SubjectID(id))
	}

	return h.getSubjectRelatedCharacters(c, id)
}

func (h Handler) getSubjectRelatedCharacters(c *fiber.Ctx, subjectID model.SubjectID) error {
	relations, err := h.c.GetSubjectRelated(c.Context(), subjectID)
	if err != nil {
		return errgo.Wrap(err, "CharacterRepo.GetSubjectRelated")
	}

	var characterIDs = make([]model.CharacterID, len(relations))
	for i, rel := range relations {
		characterIDs[i] = rel.Character.ID
	}

	var actors map[model.CharacterID][]model.Person
	if len(characterIDs) != 0 {
		actors, err = h.app.Query.GetActors(c.Context(), subjectID, characterIDs...)
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

	return c.JSON(response)
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
