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

	"github.com/bangumi/server/internal/compat"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/pkg/wiki"
)

func (h Handler) GetCharacter(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)
	id, err := req.ParseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	r, err := h.app.Query.GetCharacter(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get character", log.CharacterID(id))
	}

	if r.Redirect != 0 {
		return c.Redirect("/v0/characters/" + strconv.FormatUint(uint64(r.Redirect), 10))
	}

	return res.JSON(c, convertModelCharacter(r))
}

func (h Handler) GetCharacterComments(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)
	id, err := req.ParseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	_, err = h.app.Query.GetCharacterNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get character", log.CharacterID(id))
	}

	pagedComments, err := h.listComments(c, domain.CommentCharacter, model.TopicID(id))
	if err != nil {
		return err
	}
	return res.JSON(c, pagedComments)
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
		Gender:    null.NilString(genderMap[s.FieldGender]),
		BloodType: null.NilUint8(s.FieldBloodType),
		BirthYear: null.NilUint16(s.FieldBirthYear),
		BirthMon:  null.NilUint8(s.FieldBirthMon),
		BirthDay:  null.NilUint8(s.FieldBirthDay),
		Stat: res.Stat{
			Comments: s.CommentCount,
			Collects: s.CollectCount,
		},
		Redirect: s.Redirect,
		Locked:   s.Locked,
	}
}

func (h Handler) GetCharacterImage(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)
	id, err := req.ParseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	p, err := h.app.Query.GetCharacterNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to get character", log.CharacterID(id))
	}

	l, ok := res.PersonImage(p.Image).Select(c.Query("type"))
	if !ok {
		return res.BadRequest("bad image type: " + c.Query("type"))
	}

	if l == "" {
		return c.Redirect(res.DefaultImageURL)
	}

	return c.Redirect(l)
}

func (h Handler) GetCharacterRelatedPersons(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)
	id, err := req.ParseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	_, err = h.app.Query.GetCharacterNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to get character", log.CharacterID(id))
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
	u := h.GetHTTPAccessor(c)
	id, err := req.ParseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	_, relations, err := h.app.Query.GetCharacterRelatedSubjects(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
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
