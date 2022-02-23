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

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.PersonRepo, error) {
	return mysqlRepo{q: q, log: log.Named("subject.mysqlRepo")}, nil
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
