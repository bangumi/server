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

package timeline

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.TimeLineRepo, error) {
	return mysqlRepo{q: q, log: log.Named("timeline.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) GetByID(ctx context.Context, id model.TimeLineID) (model.TimeLine, error) {
	tl, err := m.q.TimeLine.WithContext(ctx).Where(m.q.TimeLine.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.TimeLine{}, domain.ErrNotFound
		}

		m.log.Error("unexpected happened", zap.Error(err))
		return model.TimeLine{}, errgo.Wrap(err, "dal")
	}
	return daoToModel(tl)
}

func (m mysqlRepo) ListByUID(
	ctx context.Context, uid model.UserID, limit int, since model.TimeLineID,
) ([]model.TimeLine, error) {
	tls, err := m.q.TimeLine.WithContext(ctx).
		Where(m.q.TimeLine.UID.Eq(uid), m.q.TimeLine.ID.Gt(since)).
		Order(m.q.TimeLine.Dateline).
		Limit(limit).
		Find()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []model.TimeLine{}, domain.ErrNotFound
		}
		m.log.Error("unexpected happened", zap.Error(err))
		return []model.TimeLine{}, errgo.Wrap(err, "dal")
	}
	result := make([]model.TimeLine, 0, len(tls))
	for _, tl := range tls {
		mtl, err := daoToModel(tl)
		if err != nil {
			m.log.Error("daoToModel failed", zap.Error(err))
			continue
		}
		result = append(result, mtl)
	}
	return result, nil
}

func (m mysqlRepo) Create(ctx context.Context, tls ...model.TimeLine) ([]model.TimeLine, error) {
	daoTLs := make([]*dao.TimeLine, 0, len(tls))
	convertMap := make(map[*dao.TimeLine]int, len(tls)) // record map dao->tls for id insert after create
	for i, tl := range tls {
		d, err := ModelToDAO(tl)
		if err != nil {
			return nil, errgo.Wrap(err, "modelToDAO")
		}
		daoTLs = append(daoTLs, d)
		convertMap[d] = i
	}

	if err := m.q.TimeLine.WithContext(ctx).Create(daoTLs...); err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	// reuse tls to avoid daoToModel
	for _, d := range daoTLs {
		tls[convertMap[d]].ID = d.ID
	}
	return tls, nil
}
