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

package revision

import (
	"context"
	"time"

	"go.uber.org/zap"

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

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.RevisionRepo, error) {
	return mysqlRepo{
		q,
		log,
	}, nil
}

func (r mysqlRepo) CountPersonRelated(ctx context.Context, id model.PersonIDType) (int64, error) {
	c, err := r.q.RevisionHistory.WithContext(ctx).
		Where(r.q.RevisionHistory.Type.In(model.PersonRevisionTypes()...)).Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}
	return c, nil
}

func (r mysqlRepo) ListPersonRelated(
	ctx context.Context, personID model.PersonIDType, limit int, offset int,
) ([]*model.Revision, error) {
	revisions, err := r.q.RevisionHistory.WithContext(ctx).
		Where(r.q.RevisionHistory.Type.In(model.PersonRevisionTypes()...)).
		Limit(limit).
		Offset(offset).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	result := make([]*model.Revision, len(revisions))
	for i, revision := range revisions {
		result[i] = convertRevisionDao(revision, nil)
	}
	return result, nil
}

func (r mysqlRepo) GetPersonRelated(ctx context.Context, id model.IDType) (*model.Revision, error) {
	revision, err := r.q.RevisionHistory.WithContext(ctx).
		Where(r.q.RevisionHistory.ID.Eq(id),
			r.q.RevisionHistory.Type.In(model.PersonRevisionTypes()...)).
		First()

	if err != nil {
		return &model.Revision{}, errgo.Wrap(err, "dal")
	}
	data, err := r.q.RevisionText.WithContext(ctx).
		Where(r.q.RevisionText.TextID.Eq(revision.TextID)).First()
	if err != nil {
		return &model.Revision{}, errgo.Wrap(err, "dal")
	}
	return convertRevisionDao(revision, data), nil
}

func convertRevisionDao(r *dao.RevisionHistory, data *dao.RevisionText) *model.Revision {
	var text dao.GzipPhpSerializedBlob
	if data != nil {
		text = data.Text
	}

	return &model.Revision{
		ID:        r.ID,
		Type:      r.Type,
		Summary:   r.Summary,
		CreatorID: r.CreatorID,
		CreatedAt: time.Unix(int64(r.CreatedAt), 0),
		Data:      text,
	}
}
