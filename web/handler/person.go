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
	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/pkg/vars"
	"github.com/bangumi/server/pkg/wiki"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/util"
)

func (h Handler) GetPerson(c *fiber.Ctx) error {
	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id == 0 {
		return fiber.NewError(http.StatusBadRequest, "bad id: "+strconv.Quote(c.Params("id")))
	}

	r, ok, err := h.getPersonWithCache(c.Context(), id)
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
		return c.Redirect("/v0/persons/" + strconv.FormatUint(uint64(r.Redirect), 10))
	}

	return c.JSON(r)
}

func (h Handler) getPersonWithCache(ctx context.Context, id uint32) (res.PersonV0, bool, error) {
	var key = cachekey.Person(id)

	// try to read from cache
	var r res.PersonV0
	ok, err := h.cache.Get(ctx, key, &r)
	if err != nil {
		return r, ok, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		return r, ok, nil
	}

	s, err := h.p.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.PersonV0{}, false, nil
		}

		return r, ok, errgo.Wrap(err, "repo.subject.Set")
	}

	r = convertModelPerson(s)

	if e := h.cache.Set(ctx, key, r, time.Minute); e != nil {
		logger.Error("can't set response to cache", zap.Error(e))
	}

	return r, true, nil
}

func convertModelPerson(s model.Person) res.PersonV0 {
	img := model.PersonImage(s.Image)

	return res.PersonV0{
		ID:           s.ID,
		Type:         s.Type,
		Name:         s.Name,
		Career:       careers(s),
		Images:       img,
		Summary:      s.Summary,
		LastModified: time.Time{},
		Infobox:      compat.V0Wiki(wiki.ParseOmitError(s.Infobox).NonZero()),
		Gender:       nilString(genderMap[s.FieldGender]),
		BloodType:    nilUint8(s.FieldBloodType),
		BirthYear:    nilUint16(s.FieldBirthYear),
		BirthMon:     nilUint8(s.FieldBirthMon),
		BirthDay:     nilUint8(s.FieldBirthDay),
		Stat: res.Stat{
			Comments: s.CommentCount,
			Collects: s.CollectCount,
		},
		Img:      img.Large,
		Redirect: s.Redirect,
		Locked:   s.Locked,
	}
}

//nolint:gochecknoglobals
var genderMap = map[uint8]string{
	1: "male",
	2: "female",
}

func careers(p model.Person) []string {
	s := make([]string, 0, 7)

	if p.Writer {
		s = append(s, "writer")
	}

	if p.Producer {
		s = append(s, "producer")
	}

	if p.Mangaka {
		s = append(s, "mangaka")
	}

	if p.Artist {
		s = append(s, "artist")
	}

	if p.Seiyu {
		s = append(s, "seiyu")
	}

	if p.Writer {
		s = append(s, "writer")
	}

	if p.Illustrator {
		s = append(s, "illustrator")
	}

	if p.Actor {
		s = append(s, "actor")
	}

	return s
}

func (h Handler) GetPersonRelatedCharacters(c *fiber.Ctx) error {
	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id == 0 {
		return fiber.NewError(http.StatusBadRequest, "bad id: "+c.Params("id"))
	}

	r, ok, err := h.getPersonWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	relations, err := h.c.GetPersonRelated(c.Context(), id)
	if err != nil {
		return errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	var response = make([]res.PersonRelatedCharacter, len(relations))
	for i, rel := range relations {
		response[i] = res.PersonRelatedCharacter{
			ID:            rel.Character.ID,
			Name:          rel.Character.Name,
			Type:          rel.Character.Type,
			Images:        model.PersonImage(rel.Subject.Image),
			SubjectID:     rel.Subject.ID,
			SubjectName:   rel.Subject.Name,
			SubjectNameCn: rel.Subject.NameCN,
		}
	}

	return c.JSON(response)
}

func (h Handler) GetPersonRelatedSubjects(c *fiber.Ctx) error {
	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id == 0 {
		return fiber.NewError(http.StatusBadRequest, "bad id: "+c.Params("id"))
	}

	r, ok, err := h.getPersonWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	relations, err := h.s.GetPersonRelated(c.Context(), id)
	if err != nil {
		return errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	var response = make([]res.PersonRelatedSubject, len(relations))
	for i, relation := range relations {
		response[i] = res.PersonRelatedSubject{
			SubjectID: relation.Subject.ID,
			Staff:     vars.StaffMap[relation.Subject.TypeID][relation.TypeID].String(),
			Name:      relation.Subject.Name,
			NameCn:    relation.Subject.NameCN,
			Image:     model.SubjectImage(relation.Subject.Image).Large,
		}
	}

	return c.JSON(response)
}
