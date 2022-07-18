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
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/compat"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/pkg/vars"
	"github.com/bangumi/server/pkg/wiki"
)

func (h Handler) GetPerson(c *fiber.Ctx) error {
	id, err := req.ParsePersonID(c.Params("id"))
	if err != nil {
		return err
	}

	r, err := h.app.Query.GetPerson(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get person", log.PersonID(id))
	}

	if r.Redirect != 0 {
		return c.Redirect("/v0/persons/" + strconv.FormatUint(uint64(r.Redirect), 10))
	}

	return res.JSON(c, convertModelPerson(r))
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
		Gender:       null.NilString(genderMap[s.FieldGender]),
		BloodType:    null.NilUint8(s.FieldBloodType),
		BirthYear:    null.NilUint16(s.FieldBirthYear),
		BirthMon:     null.NilUint8(s.FieldBirthMon),
		BirthDay:     null.NilUint8(s.FieldBirthDay),
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
	id, err := req.ParsePersonID(c.Params("id"))
	if err != nil {
		return err
	}

	r, err := h.app.Query.GetPerson(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get person", log.PersonID(id))
	}

	l, ok := res.PersonImage(r.Image).Select(c.Query("type"))
	if !ok {
		return res.BadRequest("bad image type: " + c.Query("type"))
	}

	if l == "" {
		return c.Redirect(res.DefaultImageURL)
	}

	return c.Redirect(l)
}

func (h Handler) GetPersonRelatedCharacters(c *fiber.Ctx) error {
	id, err := req.ParsePersonID(c.Params("id"))
	if err != nil {
		return err
	}

	r, err := h.app.Query.GetPerson(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get person", log.PersonID(id))
	}

	if r.Redirect != 0 {
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
	id, err := req.ParsePersonID(c.Params("id"))
	if err != nil {
		return err
	}

	r, err := h.app.Query.GetPerson(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get person", log.PersonID(id))
	}

	if r.Redirect != 0 {
		return res.ErrNotFound
	}

	relations, err := h.app.Query.GetPersonRelated(c.Context(), id)
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
