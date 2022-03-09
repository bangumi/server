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

package person

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/character"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/subject"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) domain.PersonRepo {
	return mysqlRepo{q: q, log: log.Named("person.mysqlRepo")}
}

func (r mysqlRepo) Get(ctx context.Context, id uint32) (model.Person, error) {
	s, err := r.q.Person.WithContext(ctx).Where(r.q.Person.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Person{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))

		return model.Person{}, errgo.Wrap(err, "dal")
	}

	field, err := r.q.PersonField.WithContext(ctx).GetPerson(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("unexpected 'gorm.ErrRecordNotFound' happened",
				zap.Error(err), zap.Uint32("id", id))

			return model.Person{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))

		return model.Person{}, errgo.Wrap(err, "dal")
	}

	return model.Person{
		Redirect:     s.Redirect,
		Type:         s.Type,
		ID:           s.ID,
		Name:         s.Name,
		Image:        s.Img,
		Infobox:      s.Infobox,
		Summary:      s.Summary,
		Locked:       s.Ban != 0,
		CollectCount: s.Collects,
		CommentCount: s.Comment,
		//
		Producer:    s.Producer,
		Mangaka:     s.Mangaka,
		Artist:      s.Artist,
		Seiyu:       s.Seiyu,
		Writer:      s.Writer,
		Illustrator: s.Illustrator,
		Actor:       s.Actor,
		//
		FieldBloodType: field.Bloodtype,
		FieldGender:    field.Gender,
		FieldBirthYear: field.BirthYear,
		FieldBirthMon:  field.BirthMon,
		FieldBirthDay:  field.BirthDay,
	}, nil
}

func (r mysqlRepo) GetSubjectRelated(
	ctx context.Context, subjectID domain.SubjectIDType,
) ([]model.Person, []model.PersonSubjectRelation, error) {
	relations, err := r.q.PersonSubjects.WithContext(ctx).
		Preload(r.q.PersonSubjects.Subject.Fields).
		Preload(r.q.PersonSubjects.Person.Fields).
		Where(r.q.PersonSubjects.SubjectID.Eq(subjectID)).Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, nil, errgo.Wrap(err, "dal")
	}

	var rel = make([]model.PersonSubjectRelation, 0, len(relations))
	var persons = make([]model.Person, 0, len(relations))
	for _, relation := range relations {
		if relation.Subject.ID == 0 || relation.Person.ID == 0 {
			// gorm/gen doesn't support preload with join, so ignore relations without subject.
			continue
		}

		rel = append(rel, model.PersonSubjectRelation{ID: relation.PrsnPosition})
		persons = append(persons, ConvertDao(&relation.Person))
	}

	return persons, rel, nil
}

func (r mysqlRepo) GetCharacterRelated(
	ctx context.Context,
	characterID domain.CharacterIDType,
) ([]domain.CharacterCast, error) {
	relations, err := r.q.Cast.WithContext(ctx).
		Preload(r.q.Cast.Character.Fields).
		Preload(r.q.Cast.Subject.Fields).
		Preload(r.q.Cast.Person.Fields).
		Where(r.q.Cast.CrtID.Eq(characterID)).
		Order(r.q.Cast.PrsnID).
		Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make([]domain.CharacterCast, 0, len(relations))
	for _, relation := range relations {
		if relation.Subject.ID == 0 || relation.Person.ID == 0 {
			// skip non-existing
			continue
		}

		results = append(results, domain.CharacterCast{
			Character: character.ConvertDao(&relation.Character),
			Person:    model.Person{},
			Subject:   subject.ConvertDao(&relation.Subject),
		})
	}

	return results, nil
}

func (r mysqlRepo) GetActors(
	ctx context.Context,
	subjectID domain.SubjectIDType,
	characterIDs ...domain.CharacterIDType,
) (map[domain.CharacterIDType][]model.Person, error) {
	relations, err := r.q.Cast.WithContext(ctx).
		Preload(r.q.Cast.Person.Fields).
		Where(r.q.Cast.CrtID.In(characterIDs...), r.q.Cast.SubjectID.Eq(subjectID)).
		Order(r.q.Cast.PrsnID).
		Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make(map[domain.CharacterIDType][]model.Person, len(relations))
	for _, relation := range relations {
		if relation.Person.ID == 0 {
			// skip non-existing
			continue
		}

		// should pre-alloc a big slice and split it as results.
		results[relation.CrtID] = append(results[relation.Person.ID], ConvertDao(&relation.Person))
	}

	return results, nil
}

func ConvertDao(p *dao.Person) model.Person {
	return model.Person{
		Redirect:     p.Redirect,
		Type:         p.Type,
		ID:           p.ID,
		Name:         p.Name,
		Image:        p.Img,
		Infobox:      p.Infobox,
		Summary:      p.Summary,
		Locked:       p.Ban != 0,
		CollectCount: p.Collects,
		CommentCount: p.Comment,
		//
		Producer:    p.Producer,
		Mangaka:     p.Mangaka,
		Artist:      p.Artist,
		Seiyu:       p.Seiyu,
		Writer:      p.Writer,
		Illustrator: p.Illustrator,
		Actor:       p.Actor,
		//
		FieldBloodType: p.Fields.Bloodtype,
		FieldGender:    p.Fields.Gender,
		FieldBirthYear: p.Fields.BirthYear,
		FieldBirthMon:  p.Fields.BirthMon,
		FieldBirthDay:  p.Fields.BirthDay,
	}
}
