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

package character

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

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

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.CharacterRepo, error) {
	return mysqlRepo{q: q, log: log.Named("character.mysqlRepo")}, nil
}

func (r mysqlRepo) Get(ctx context.Context, id uint32) (model.Character, error) {
	s, err := r.q.Character.WithContext(ctx).Where(r.q.Character.ID.Eq(id)).Limit(1).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Character{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))

		return model.Character{}, errgo.Wrap(err, "dal")
	}

	field, err := r.q.PersonField.WithContext(ctx).GetCharacter(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("unexpected 'gorm.ErrRecordNotFound' happened",
				zap.Error(err), zap.Uint32("id", id))

			return model.Character{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))

		return model.Character{}, errgo.Wrap(err, "dal")
	}

	return model.Character{
		ID:           s.ID,
		Name:         s.Name,
		Type:         s.Role,
		Image:        s.Img,
		Summary:      s.Summary,
		Locked:       s.Ban != 0,
		Infobox:      s.Infobox,
		CollectCount: s.Collects,
		CommentCount: s.Comment,
		NSFW:         s.Nsfw,
		//
		FieldBloodType: field.Bloodtype,
		FieldGender:    field.Gender,
		FieldBirthYear: field.BirthYear,
		FieldBirthMon:  field.BirthMon,
		FieldBirthDay:  field.BirthDay,
		//
		Redirect: s.Redirect,
	}, nil
}

func (r mysqlRepo) GetPersonRelated(
	ctx context.Context, characterID domain.PersonIDType,
) ([]model.Character, []model.Subject, []model.PersonCharacterRelation, error) {
	relations, err := r.q.CharacterSubjects.WithContext(ctx).
		Preload(r.q.CharacterSubjects.Character.Fields).Preload(r.q.CharacterSubjects.Subject.Fields).
		Where(r.q.CharacterSubjects.CrtID.Eq(characterID)).Order(r.q.CharacterSubjects.CrtID).
		Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, nil, nil, errgo.Wrap(err, "dal")
	}

	var rel = make([]model.PersonCharacterRelation, 0, len(relations))
	var characters = make([]model.Character, 0, len(relations))
	var subjects = make([]model.Subject, 0, len(relations))

	for _, relation := range relations {
		if relation.Subject.ID == 0 || relation.Character.ID == 0 {
			// skip non-existing
			continue
		}

		rel = append(rel, model.PersonCharacterRelation{Type: relation.CrtType})
		characters = append(characters, convertDao(&relation.Character))
		subjects = append(subjects, subject.ConvertDao(&relation.Subject))
	}

	return characters, subjects, rel, nil
}

func convertDao(s *dao.Character) model.Character {
	return model.Character{
		ID:           s.ID,
		Name:         s.Name,
		Type:         s.Role,
		Image:        s.Img,
		Summary:      s.Summary,
		Locked:       s.Ban != 0,
		Infobox:      s.Infobox,
		CollectCount: s.Collects,
		CommentCount: s.Comment,
		NSFW:         s.Nsfw,
		//
		FieldBloodType: s.Fields.Bloodtype,
		FieldGender:    s.Fields.Gender,
		FieldBirthYear: s.Fields.BirthYear,
		FieldBirthMon:  s.Fields.BirthMon,
		FieldBirthDay:  s.Fields.BirthDay,
		//
		Redirect: s.Redirect,
	}
}
