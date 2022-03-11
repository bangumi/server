// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/compat"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/pkg/vars"
	"github.com/bangumi/server/pkg/wiki"
	"github.com/bangumi/server/web/handler/cachekey"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/util"
)

func (h Handler) GetSubject(c *fiber.Ctx) error {
	u := h.getUser(c)

	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id == 0 {
		return fiber.NewError(http.StatusBadRequest, "bad id: "+c.Params("id"))
	}

	r, ok, err := h.getSubjectWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	if r.Redirect != 0 {
		return c.Redirect("/v0/subjects/" + strconv.FormatUint(uint64(r.Redirect), 10))
	}

	if r.NSFW && !u.AllowNSFW() {
		// default Handler will return a 404 response
		return c.Next()
	}

	return c.JSON(r)
}

func (h Handler) getSubjectWithCache(
	ctx context.Context,
	id domain.SubjectIDType,
) (res.SubjectV0, bool, error) {
	var key = cachekey.Subject(id)

	// try to read from cache
	var r res.SubjectV0
	ok, err := h.cache.Get(ctx, key, &r)
	if err != nil {
		return r, ok, errgo.Wrap(err, "cache.Get")
	}

	if !ok {
		return r, ok, nil
	}

	s, err := h.s.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.SubjectV0{}, false, nil
		}

		return r, ok, errgo.Wrap(err, "repo.subject.Set")
	}

	r = convertModelSubject(s)
	r.TotalEpisodes, err = h.e.Count(ctx, id)
	if err != nil {
		return r, false, errgo.Wrap(err, "repo.episode.Count")
	}

	if e := h.cache.Set(ctx, key, r, time.Minute); e != nil {
		logger.Error("can't set response to cache", zap.Error(e))
	}

	return r, true, nil
}

func platformString(s model.Subject) *string {
	platform, ok := vars.PlatformMap[s.TypeID][s.PlatformID]
	if ok {
		logger.Warn("unknown platform",
			zap.Uint32("subject_id", s.ID),
			zap.Uint8("type", s.TypeID),
			zap.Uint16("platform", s.PlatformID),
		)

		return nil
	}

	v := platform.String()

	return &v
}

func (h Handler) GetSubjectRelatedPersons(c *fiber.Ctx) error {
	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id == 0 {
		return fiber.NewError(http.StatusBadRequest, "bad id: "+strconv.Quote(c.Params("id")))
	}

	r, ok, err := h.getSubjectWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	persons, relations, err := h.p.GetSubjectRelated(c.Context(), id)
	if err != nil {
		return errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	var response = make([]res.SubjectRelatedPerson, len(persons))
	for i, p := range persons {
		response[i] = res.SubjectRelatedPerson{
			Images:   model.PersonImage(p.Image),
			Name:     p.Name,
			Relation: vars.StaffMap[r.TypeID][relations[i].ID].String(),
			Career:   careers(p),
			Type:     p.Type,
			ID:       p.ID,
		}
	}

	return c.JSON(response)
}

func convertModelSubject(s model.Subject) res.SubjectV0 {
	tags, err := compat.ParseTags(s.CompatRawTags)
	if err != nil {
		logger.Warn("failed to parse tags", zap.Uint32("subject_id", s.ID))
	}

	var date *string
	if s.Date != "" {
		date = &s.Date
	}

	return res.SubjectV0{
		ID:       s.ID,
		Image:    model.SubjectImage(s.Image),
		Summary:  s.Summary,
		Name:     s.Name,
		Platform: platformString(s),
		NameCN:   s.NameCN,
		Date:     date,
		Infobox:  compat.V0Wiki(wiki.ParseOmitError(s.Infobox).NonZero()),
		Volumes:  s.Volumes,
		Redirect: s.Redirect,
		Eps:      s.Eps,
		Tags:     tags,
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
	u := h.getUser(c)

	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id == 0 {
		return fiber.NewError(http.StatusBadRequest, "bad id: "+strconv.Quote(c.Params("id")))
	}

	r, ok, err := h.getSubjectWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 || (r.NSFW && !u.AllowNSFW()) {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	subjects, relation, err := h.s.GetSubjectRelated(c.Context(), id)
	if err != nil {
		return errgo.Wrap(err, "repo")
	}

	var response = make([]res.SubjectRelatedSubject, len(subjects))
	for i, subject := range subjects {
		response[i] = res.SubjectRelatedSubject{
			Images:    model.SubjectImage(subject.Image),
			Name:      subject.Name,
			NameCn:    subject.NameCN,
			Relation:  readableRelation(subject.TypeID, relation[i].Type),
			Type:      subject.TypeID,
			SubjectID: subject.ID,
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
	u := h.getUser(c)
	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id == 0 {
		return fiber.NewError(http.StatusBadRequest, "bad id: "+strconv.Quote(c.Params("id")))
	}

	r, ok, err := h.getSubjectWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 || (r.NSFW && !u.AllowNSFW()) {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	return h.getSubjectRelatedCharacters(c, id)
}

func (h Handler) getSubjectRelatedCharacters(c *fiber.Ctx, subjectID domain.SubjectIDType) error {
	characters, relations, err := h.c.GetSubjectRelated(c.Context(), subjectID)
	if err != nil {
		return errgo.Wrap(err, "CharacterRepo.GetSubjectRelated")
	}

	var characterIDs = make([]domain.PersonIDType, len(characters))
	for i, character := range characters {
		characterIDs[i] = character.ID
	}

	actors, err := h.p.GetActors(c.Context(), subjectID, characterIDs...)
	if err != nil {
		return errgo.Wrap(err, "PersonRepo.GetActors")
	}

	var response = make([]res.SubjectRelatedCharacter, len(characters))
	for i, character := range characters {
		response[i] = res.SubjectRelatedCharacter{
			Images:   model.PersonImage(character.Image),
			Name:     character.Name,
			Relation: characterStaffString(relations[i].Type),
			Actors:   toActors(actors[character.ID]),
			Type:     character.Type,
			ID:       character.ID,
		}
	}

	return c.JSON(response)
}

func toActors(persons []model.Person) []res.Actor {
	// should pre-alloc a big slice and split it into sub slice.
	var actors = make([]res.Actor, len(persons))
	for j, actor := range persons {
		actors[j] = res.Actor{
			Images:       model.PersonImage(actor.Image),
			Name:         actor.Name,
			ShortSummary: actor.Summary,
			Career:       careers(actor),
			ID:           actor.ID,
			Type:         actor.Type,
			Locked:       actor.Locked,
		}
	}

	return actors
}
