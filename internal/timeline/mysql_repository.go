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
	"bytes"
	"context"
	"errors"

	"github.com/bangumi/server/internal/web/res"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/timeline/image"
	"github.com/bangumi/server/internal/timeline/memo"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.TimeLineRepo, error) {
	return mysqlRepo{q: q, log: log.Named("timeline.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) GetByID(ctx context.Context, id model.TimeLineID) (*model.TimeLine, error) {
	tl, err := m.q.TimeLine.WithContext(ctx).Where(m.q.TimeLine.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		m.log.Error("unexpected happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}
	return daoToModel(tl)
}

func (m mysqlRepo) ListByUID(
	ctx context.Context, uid model.UserID, limit int, since model.TimeLineID,
) ([]*model.TimeLine, error) {
	tls, err := m.q.TimeLine.WithContext(ctx).
		Where(m.q.TimeLine.UID.Eq(uid), m.q.TimeLine.ID.Gt(since)).
		Order(m.q.TimeLine.Dateline).
		Limit(limit).
		Find()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		m.log.Error("unexpected happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}
	result := make([]*model.TimeLine, 0, len(tls))
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

func (m mysqlRepo) Create(ctx context.Context, tl *model.TimeLine) error {
	d, err := modelToDAO(tl)
	if err != nil {
		m.log.Error("modelToDAO", zap.Error(err))
		return errgo.Wrap(err, "modelToDAO")
	}

	isDuped, err := m.isDupeTimeLine(ctx, d)
	if err != nil {
		m.log.Error("isDupeTimeLine", zap.Error(err))
		return errgo.Wrap(err, "isDupeTimeLine")
	}
	if isDuped {
		return res.BadRequest("duplicated timeline")
	}

	if err := m.q.TimeLine.WithContext(ctx).Create(d); err != nil {
		return errgo.Wrap(err, "dal")
	}
	return nil
}

func (m mysqlRepo) isDupeTimeLine(ctx context.Context, dao *dao.TimeLine) (bool, error) {
	daoTLs, err := m.q.TimeLine.WithContext(ctx).
		Where(m.q.TimeLine.UID.Eq(dao.UID)).
		Order(m.q.TimeLine.Dateline.Desc()).
		Limit(1).
		Find()
	if err != nil {
		return false, errgo.Wrap(err, "dal")
	}

	if len(daoTLs) == 0 {
		return false, nil
	}
	daoTL := daoTLs[0]
	if dao.Dateline != daoTL.Dateline {
		return false, nil
	}
	if dao.Cat != daoTL.Cat {
		return false, nil
	}
	if bytes.Compare(dao.Memo, daoTL.Memo) != 0 {
		return false, nil
	}
	if bytes.Compare(dao.Img, daoTL.Img) != 0 {
		return false, nil
	}
	return true, nil
}

func daoToModel(tl *dao.TimeLine) (*model.TimeLine, error) {
	mm, err := memo.DAOToModel(tl)
	if err != nil {
		return nil, errgo.Wrap(err, "DAOToModel")
	}

	img, err := image.DAOToModel(tl)
	if err != nil {
		return nil, errgo.Wrap(err, "DAOToModel")
	}

	return &model.TimeLine{
		ID:       tl.ID,
		Related:  tl.Related,
		Memo:     *mm,
		Image:    img,
		UID:      tl.UID,
		Replies:  tl.Replies,
		Dateline: tl.Dateline,
		Cat:      tl.Cat,
		Type:     tl.Type,
		Batch:    tl.Batch,
		Source:   tl.Source,
	}, nil
}

func modelToDAO(tl *model.TimeLine) (*dao.TimeLine, error) {
	img, err := image.ModelToDAO(tl)
	if err != nil {
		return nil, errgo.Wrap(err, "modelImageToDAO")
	}

	mm, err := memo.ModelToDAO(tl)
	if err != nil {
		return nil, errgo.Wrap(err, "ModelToDAO")
	}

	return &dao.TimeLine{
		ID:       tl.ID,
		Related:  tl.Related,
		Img:      img,
		Memo:     mm,
		UID:      tl.UID,
		Replies:  tl.Replies,
		Dateline: tl.Dateline,
		Cat:      tl.Cat,
		Type:     tl.Type,
		Batch:    tl.Batch,
		Source:   tl.Source,
	}, nil
}
