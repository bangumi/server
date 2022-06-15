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

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) domain.PersonRepo {
	return mysqlRepo{q: q, log: log.Named("person.mysqlRepo")}
}

func (r mysqlRepo) Get(ctx context.Context, id uint32) (model.Person, error) {
	p, err := r.q.Person.WithContext(ctx).Joins(r.q.Person.Fields).Where(r.q.Person.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Person{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))

		return model.Person{}, errgo.Wrap(err, "dal")
	}

	return ConvertDao(p), nil
}

func (r mysqlRepo) GetSubjectRelated(
	ctx context.Context, subjectID model.SubjectID,
) ([]domain.SubjectPersonRelation, error) {
	relations, err := r.q.PersonSubjects.WithContext(ctx).
		Where(r.q.PersonSubjects.SubjectID.Eq(subjectID)).Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var rel = make([]domain.SubjectPersonRelation, len(relations))
	for i, relation := range relations {
		rel[i] = domain.SubjectPersonRelation{
			PersonID: relation.PersonID,
			TypeID:   relation.PrsnPosition,
		}
	}

	return rel, nil
}

func (r mysqlRepo) GetCharacterRelated(
	ctx context.Context,
	characterID model.CharacterID,
) ([]domain.PersonCharacterRelation, error) {
	relations, err := r.q.Cast.WithContext(ctx).
		Where(r.q.Cast.CharacterID.Eq(characterID)).
		Order(r.q.Cast.PersonID).
		Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make([]domain.PersonCharacterRelation, 0, len(relations))
	for _, relation := range relations {
		results = append(results, domain.PersonCharacterRelation{
			CharacterID: relation.CharacterID,
			PersonID:    relation.PersonID,
			SubjectID:   relation.SubjectID,
		})
	}

	return results, nil
}

func (r mysqlRepo) GetByIDs(
	ctx context.Context, ids ...model.PersonID,
) (map[model.PersonID]model.Person, error) {
	u, err := r.q.Person.WithContext(ctx).Joins(r.q.Person.Fields).Where(r.q.Person.ID.In(ids...)).Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var result = make(map[model.PersonID]model.Person, len(ids))
	for _, p := range u {
		result[p.ID] = ConvertDao(p)
	}

	return result, nil
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
