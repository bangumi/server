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
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.CharacterRepo, error) {
	return mysqlRepo{q: q, log: log.Named("character.mysqlRepo")}, nil
}

func (r mysqlRepo) Get(ctx context.Context, id uint32) (model.Character, error) {
	s, err := r.q.Character.WithContext(ctx).Preload(r.q.Character.Fields).Where(r.q.Character.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Character{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return model.Character{}, errgo.Wrap(err, "dal")
	}

	return ConvertDao(s), nil
}

func (r mysqlRepo) GetByIDs(
	ctx context.Context, ids ...model.CharacterIDType,
) (map[model.CharacterIDType]model.Character, error) {
	records, err := r.q.Character.WithContext(ctx).Preload(r.q.Character.Fields).Where(r.q.Character.ID.In(ids...)).Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make(map[model.CharacterIDType]model.Character, len(records))
	for _, s := range records {
		results[s.ID] = ConvertDao(s)
	}

	return results, nil
}

func (r mysqlRepo) GetPersonRelated(
	ctx context.Context, personID model.PersonIDType,
) ([]domain.PersonCharacterRelation, error) {
	relations, err := r.q.Cast.WithContext(ctx).
		Where(r.q.Cast.PersonID.Eq(personID)).
		Order(r.q.Cast.SubjectID).
		Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var rel = make([]domain.PersonCharacterRelation, 0, len(relations))
	for _, relation := range relations {
		rel = append(rel, domain.PersonCharacterRelation{
			CharacterID: relation.PersonID,
			SubjectID:   relation.SubjectID,
			PersonID:    personID,
		})
	}

	return rel, nil
}

func (r mysqlRepo) GetSubjectRelated(
	ctx context.Context, subjectID model.SubjectIDType,
) ([]domain.SubjectCharacterRelation, error) {
	relations, err := r.q.CharacterSubjects.WithContext(ctx).
		Where(r.q.CharacterSubjects.SubjectID.Eq(subjectID)).
		Order(r.q.CharacterSubjects.CharacterID).Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var rel = make([]domain.SubjectCharacterRelation, len(relations))
	for i, relation := range relations {
		rel[i] = domain.SubjectCharacterRelation{
			CharacterID: relation.CharacterID,
			SubjectID:   relation.SubjectID,
			TypeID:      relation.CrtType,
		}
	}

	return rel, nil
}

func ConvertDao(s *dao.Character) model.Character {
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
