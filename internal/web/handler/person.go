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
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/compat"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/pkg/vars"
	"github.com/bangumi/server/pkg/wiki"
)

func (h Handler) GetPerson(c *fiber.Ctx) error {
	id, err := parsePersonID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getPersonWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok {
		return res.ErrNotFound
	}

	if r.Redirect != 0 {
		return c.Redirect("/v0/persons/" + strconv.FormatUint(uint64(r.Redirect), 10))
	}

	return c.JSON(r)
}

// first try to read from cache, then fallback to reading from database.
// return data, database record existence and error.
func (h Handler) getPersonWithCache(ctx context.Context, id model.PersonID) (res.PersonV0, bool, error) {
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

		return r, ok, errgo.Wrap(err, "personRepo.Get")
	}

	r = convertModelPerson(s)

	if e := h.cache.Set(ctx, key, r, time.Minute); e != nil {
		logger.Error("can't set response to cache", zap.Error(e))
	}

	return r, true, nil
}

func convertModelPerson(s model.Person) res.PersonV0 {
	img := res.PersonImage(s.Image)

	return res.PersonV0{
		ID:           s.ID,
		Type:         s.Type,
		Name:         s.Name,
		Career:       s.Careers(),
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

func (h Handler) GetPersonImage(c *fiber.Ctx) error {
	id, err := parsePersonID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getPersonWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok {
		return res.ErrNotFound
	}

	l, ok := r.Images.Select(c.Query("type"))
	if !ok {
		return res.BadRequest("bad image type: " + c.Query("type"))
	}

	if l == "" {
		return c.Redirect(res.DefaultImageURL)
	}

	return c.Redirect(l)
}

func (h Handler) GetPersonRelatedCharacters(c *fiber.Ctx) error {
	id, err := parsePersonID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getPersonWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 {
		return res.ErrNotFound
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
			Images:        res.PersonImage(rel.Subject.Image),
			SubjectID:     rel.Subject.ID,
			SubjectName:   rel.Subject.Name,
			SubjectNameCn: rel.Subject.NameCN,
		}
	}

	return c.JSON(response)
}

func (h Handler) GetPersonRelatedSubjects(c *fiber.Ctx) error {
	id, err := parsePersonID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getPersonWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 {
		return res.ErrNotFound
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
			Image:     res.SubjectImage(relation.Subject.Image).Large,
		}
	}

	return c.JSON(response)
}
