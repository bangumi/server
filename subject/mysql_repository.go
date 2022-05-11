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

package subject

import (
	"context"
	"errors"
	"math"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/person"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.SubjectRepo, error) {
	return mysqlRepo{q: q, log: log.Named("subject.mysqlRepo")}, nil
}

func (r mysqlRepo) Get(ctx context.Context, id uint32) (model.Subject, error) {
	s, err := r.q.Subject.WithContext(ctx).Preload(r.q.Subject.Fields).Where(r.q.Subject.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Subject{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return model.Subject{}, errgo.Wrap(err, "dal")
	}

	return ConvertDao(s), nil
}

func ConvertDao(s *dao.Subject) model.Subject {
	var date string
	if !s.Fields.Date.IsZero() {
		date = s.Fields.Date.Format("2006-01-02")
	}

	return model.Subject{
		Redirect:      s.Fields.Redirect,
		Date:          date,
		ID:            s.ID,
		Name:          s.Name,
		NameCN:        s.NameCN,
		TypeID:        s.TypeID,
		Image:         s.Image,
		PlatformID:    s.Platform,
		Infobox:       s.Infobox,
		Summary:       s.Summary,
		Volumes:       s.Volumes,
		Eps:           s.Eps,
		Wish:          s.Wish,
		Collect:       s.Collect,
		Doing:         s.Doing,
		OnHold:        s.OnHold,
		CompatRawTags: s.Fields.Tags,
		Dropped:       s.Dropped,
		Airtime:       s.Airtime,
		NSFW:          s.Nsfw,
		Ban:           s.Ban,
		Rating:        rating(s.Fields),
	}
}

func rating(f dao.SubjectField) model.Rating {
	var total = f.Rate1 + f.Rate2 + f.Rate3 + f.Rate4 + f.Rate5 +
		f.Rate6 + f.Rate7 + f.Rate8 + f.Rate9 + f.Rate10

	if total == 0 {
		return model.Rating{}
	}

	var score = float64(1*f.Rate1+2*f.Rate2+3*f.Rate3+4*f.Rate4+5*f.Rate5+
		6*f.Rate6+7*f.Rate7+8*f.Rate8+9*f.Rate9+10*f.Rate10) / float64(total)

	return model.Rating{
		Rank:  f.Rank,
		Total: total,
		Count: model.Count{
			Field1:  f.Rate1,
			Field2:  f.Rate2,
			Field3:  f.Rate3,
			Field4:  f.Rate4,
			Field5:  f.Rate5,
			Field6:  f.Rate6,
			Field7:  f.Rate7,
			Field8:  f.Rate8,
			Field9:  f.Rate9,
			Field10: f.Rate10,
		},
		Score: math.Round(score*10) / 10,
	}
}

func (r mysqlRepo) GetPersonRelated(
	ctx context.Context, personID model.PersonIDType,
) ([]domain.SubjectPersonRelation, error) {
	relations, err := r.q.PersonSubjects.WithContext(ctx).
		Joins(r.q.PersonSubjects.Subject).
		Joins(r.q.PersonSubjects.Person).
		Where(r.q.PersonSubjects.PersonID.Eq(personID)).Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))

		return nil, errgo.Wrap(err, "dal")
	}

	var rel = make([]domain.SubjectPersonRelation, 0, len(relations))
	for _, relation := range relations {
		rel = append(rel, domain.SubjectPersonRelation{
			SubjectID: relation.SubjectID,
			PersonID:  relation.PersonID,
			TypeID:    relation.PrsnPosition,
		})
	}

	return rel, nil
}

func (r mysqlRepo) GetCharacterRelated(
	ctx context.Context,
	characterID model.PersonIDType,
) ([]domain.SubjectCharacterRelation, error) {
	relations, err := r.q.CharacterSubjects.WithContext(ctx).
		Joins(r.q.CharacterSubjects.Subject.Fields).
		Where(r.q.CharacterSubjects.CharacterID.Eq(characterID)).Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var rel = make([]domain.SubjectCharacterRelation, 0, len(relations))
	for _, relation := range relations {
		rel = append(rel, domain.SubjectCharacterRelation{
			SubjectID:   relation.Subject.ID,
			CharacterID: relation.CharacterID,
			TypeID:      relation.CrtType,
		})
	}

	return rel, nil
}

func (r mysqlRepo) GetSubjectRelated(
	ctx context.Context,
	subjectID model.SubjectIDType,
) ([]domain.SubjectInternalRelation, error) {
	relations, err := r.q.SubjectRelation.WithContext(ctx).
		Joins(r.q.SubjectRelation.Subject).Where(r.q.SubjectRelation.SubjectID.Eq(subjectID)).
		Order(r.q.SubjectRelation.Order).Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var rel = make([]domain.SubjectInternalRelation, 0, len(relations))
	for _, relation := range relations {
		rel = append(rel, domain.SubjectInternalRelation{
			SourceID:      subjectID,
			DestinationID: relation.Subject.ID,
			TypeID:        relation.RelationType,
		})
	}

	return rel, nil
}

func (r mysqlRepo) GetByIDs(
	ctx context.Context, ids ...model.SubjectIDType,
) (map[model.SubjectIDType]model.Subject, error) {
	records, err := r.q.Subject.WithContext(ctx).Joins(r.q.Subject.Fields).Where(r.q.Subject.ID.In(ids...)).Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var result = make(map[model.SubjectIDType]model.Subject, len(ids))

	for _, s := range records {
		result[s.ID] = ConvertDao(s)
	}

	return result, nil
}
func (r mysqlRepo) GetActors(
	ctx context.Context,
	subjectID model.SubjectIDType,
	characterIDs ...model.CharacterIDType,
) (map[model.CharacterIDType][]model.Person, error) {
	relations, err := r.q.Cast.WithContext(ctx).
		Preload(r.q.Cast.Person.Fields).
		Where(r.q.Cast.CharacterID.In(characterIDs...), r.q.Cast.SubjectID.Eq(subjectID)).
		Order(r.q.Cast.PersonID).
		Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make(map[model.CharacterIDType][]model.Person, len(relations))
	for _, relation := range relations {
		// TODO: should pre-alloc a big slice and split it as results.
		results[relation.CharacterID] = append(results[relation.CharacterID], person.ConvertDao(&relation.Person))
	}

	return results, nil
}
