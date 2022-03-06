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

package index

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.IndexRepo, error) {
	return mysqlRepo{q: q, log: log.Named("index.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (r mysqlRepo) IsNsfw(ctx context.Context, id uint32) (bool, error) {
	i, err := r.q.IndexSubject.WithContext(ctx).
		Join(r.q.Subject, r.q.IndexSubject.Sid.EqCol(r.q.Subject.ID)).
		Where(r.q.IndexSubject.Rid.Eq(id), r.q.Subject.Nsfw.Is(true)).Count()
	if err != nil {
		return false, errgo.Wrap(err, "dal")
	}

	return i == 0, nil
}

func (r mysqlRepo) Get(ctx context.Context, id uint32) (model.Index, error) {
	i, err := r.q.Index.WithContext(ctx).
		Where(r.q.Index.ID.Eq(id), r.q.Index.Ban.Is(false)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Index{}, domain.ErrNotFound
		}

		return model.Index{}, errgo.Wrap(err, "dal")
	}

	return model.Index{
		CreatedAt:   time.Unix(int64(i.Dateline), 0),
		Title:       i.Title,
		Description: i.Desc,
		CreatorID:   i.UID,
		Total:       i.SubjectTotal,
		ID:          id,
		Comments:    i.Replies,
		Collects:    i.Collects,
		Ban:         i.Ban,
	}, nil
}
