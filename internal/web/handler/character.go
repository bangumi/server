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

	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/compat"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/util"
	"github.com/bangumi/server/pkg/wiki"
)

func (h Handler) GetCharacter(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)
	id, err := parseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getCharacterWithCache(c.Context(), id)
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
		return c.Redirect("/v0/characters/" + strconv.FormatUint(uint64(r.Redirect), 10))
	}

	if r.NSFW && !u.AllowNSFW() {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	return c.JSON(r)
}

// first try to read from cache, then fallback to reading from database.
// return data, database record existence and error.
func (h Handler) getCharacterWithCache(
	ctx context.Context, id model.CharacterID) (res.CharacterV0, bool, error) {
	var key = cachekey.Character(id)

	// try to read from cache
	var r res.CharacterV0
	ok, err := h.cache.Get(ctx, key, &r)
	if err != nil {
		return r, ok, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		return r, ok, nil
	}

	s, err := h.c.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.CharacterV0{}, false, nil
		}

		return r, ok, errgo.Wrap(err, "repo.subject.Set")
	}

	r = convertModelCharacter(s)

	if e := h.cache.Set(ctx, key, r, time.Minute); e != nil {
		logger.Error("can't set response to cache", zap.Error(e))
	}

	return r, true, nil
}

func convertModelCharacter(s model.Character) res.CharacterV0 {
	img := res.PersonImage(s.Image)

	return res.CharacterV0{
		ID:        s.ID,
		Type:      s.Type,
		Name:      s.Name,
		NSFW:      s.NSFW,
		Images:    img,
		Summary:   s.Summary,
		Infobox:   compat.V0Wiki(wiki.ParseOmitError(s.Infobox).NonZero()),
		Gender:    nilString(genderMap[s.FieldGender]),
		BloodType: nilUint8(s.FieldBloodType),
		BirthYear: nilUint16(s.FieldBirthYear),
		BirthMon:  nilUint8(s.FieldBirthMon),
		BirthDay:  nilUint8(s.FieldBirthDay),
		Stat: res.Stat{
			Comments: s.CommentCount,
			Collects: s.CollectCount,
		},
		Redirect: s.Redirect,
		Locked:   s.Locked,
	}
}

func (h Handler) GetCharacterImage(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)
	id, err := parseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getCharacterWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.NSFW && !u.AllowNSFW() {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	l, ok := r.Images.Select(c.Query("type"))
	if !ok {
		return fiber.NewError(http.StatusBadRequest, "bad image type: "+c.Query("type"))
	}

	if l == "" {
		return c.Redirect(res.DefaultImageURL)
	}

	return c.Redirect(l)
}

func (h Handler) GetCharacterRelatedPersons(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)
	id, err := parseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getCharacterWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 || r.NSFW && !u.AllowNSFW() {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	casts, err := h.p.GetCharacterRelated(c.Context(), id)
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

	return c.JSON(response)
}

func (h Handler) GetCharacterRelatedSubjects(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)
	id, err := parseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getCharacterWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 || (r.NSFW && !u.AllowNSFW()) {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	relations, err := h.s.GetCharacterRelated(c.Context(), id)
	if err != nil {
		return errgo.Wrap(err, "repo.GetCharacterRelated")
	}

	var response = make([]res.CharacterRelatedSubject, len(relations))
	for i, relation := range relations {
		subject := relation.Subject
		response[i] = res.CharacterRelatedSubject{
			ID:     subject.ID,
			Name:   subject.Name,
			NameCn: subject.NameCN,
			Staff:  characterStaffString(relation.TypeID),
			Image:  res.SubjectImage(subject.Image).Large,
		}
	}

	return c.JSON(response)
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
