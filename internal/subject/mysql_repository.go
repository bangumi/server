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
	"fmt"
	"math"

	"github.com/trim21/errgo"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (Repo, error) {
	return mysqlRepo{q: q, log: log.Named("subject.mysqlRepo")}, nil
}

func (r mysqlRepo) Get(ctx context.Context, id model.SubjectID, filter Filter) (model.Subject, error) {
	q := r.q.Subject.WithContext(ctx).Preload(r.q.Subject.Fields).Where(r.q.Subject.ID.Eq(id))

	if filter.NSFW.Set {
		q = q.Where(r.q.Subject.Nsfw.Is(filter.NSFW.Value))
	}

	s, err := q.Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Subject{}, fmt.Errorf("%w: %d", gerr.ErrNotFound, id)
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return model.Subject{}, errgo.Wrap(err, "dal")
	}

	return ConvertDao(s)
}

func ConvertDao(s *dao.Subject) (model.Subject, error) {
	var date string
	if !s.Fields.Date.IsZero() {
		date = s.Fields.Date.Format("2006-01-02")
	}

	tags, err := ParseTags(s.Fields.Tags)
	if err != nil {
		return model.Subject{}, err
	}

	return model.Subject{
		Redirect:   s.Fields.Redirect,
		Date:       date,
		ID:         s.ID,
		Name:       s.Name,
		NameCN:     s.NameCN,
		TypeID:     s.TypeID,
		Image:      s.Image,
		PlatformID: s.Platform,
		Infobox:    s.Infobox,
		Summary:    s.Summary,
		Volumes:    s.Volumes,
		Eps:        s.Eps,
		Wish:       s.Wish,
		Collect:    s.Done,
		Doing:      s.Doing,
		OnHold:     s.OnHold,
		Series:     s.Series,
		Tags:       tags,
		Dropped:    s.Dropped,
		Airtime:    s.Airtime,
		NSFW:       s.Nsfw,
		Ban:        s.Ban,
		Rating:     rating(s.Fields),
	}, nil
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
	ctx context.Context, personID model.PersonID,
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
	characterID model.CharacterID,
) ([]domain.SubjectCharacterRelation, error) {
	relations, err := r.q.CharacterSubjects.WithContext(ctx).
		Joins(r.q.CharacterSubjects.Subject).
		Where(r.q.CharacterSubjects.CharacterID.Eq(characterID)).Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var rel = make([]domain.SubjectCharacterRelation, 0, len(relations))
	for _, relation := range relations {
		rel = append(rel, domain.SubjectCharacterRelation{
			SubjectID:   relation.SubjectID,
			CharacterID: relation.CharacterID,
			TypeID:      relation.CrtType,
		})
	}

	return rel, nil
}

func (r mysqlRepo) GetSubjectRelated(
	ctx context.Context,
	subjectID model.SubjectID,
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
	ctx context.Context, ids []model.SubjectID, filter Filter,
) (map[model.SubjectID]model.Subject, error) {
	if len(ids) == 0 {
		return map[model.SubjectID]model.Subject{}, nil
	}
	q := r.q.Subject.WithContext(ctx).Joins(r.q.Subject.Fields).Where(r.q.Subject.ID.In(ids...))

	if filter.NSFW.Set {
		q = q.Where(r.q.Subject.Nsfw.Is(filter.NSFW.Value))
	}

	records, err := q.Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var result = make(map[model.SubjectID]model.Subject, len(ids))

	for _, s := range records {
		result[s.ID], err = ConvertDao(s)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r mysqlRepo) GetActors(
	ctx context.Context,
	subjectID model.SubjectID,
	characterIDs []model.CharacterID,
) (map[model.CharacterID][]model.PersonID, error) {
	relations, err := r.q.Cast.WithContext(ctx).
		Where(r.q.Cast.CharacterID.In(characterIDs...), r.q.Cast.SubjectID.Eq(subjectID)).
		Order(r.q.Cast.PersonID).
		Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make(map[model.CharacterID][]model.PersonID, len(relations))
	for _, relation := range relations {
		// TODO: should pre-alloc a big slice and split it as results.
		results[relation.CharacterID] = append(results[relation.CharacterID], relation.PersonID)
	}

	return results, nil
}
