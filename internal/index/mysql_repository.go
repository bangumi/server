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

	"github.com/trim21/errgo"
	"go.uber.org/zap"
	"gorm.io/gen"
	"gorm.io/gorm"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/subject"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (Repo, error) {
	return mysqlRepo{q: q, log: log.Named("index.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (r mysqlRepo) isNsfw(ctx context.Context, id model.IndexID) (bool, error) {
	i, err := r.q.IndexSubject.WithContext(ctx).
		Join(r.q.Subject, r.q.IndexSubject.SubjectID.EqCol(r.q.Subject.ID)).
		Where(r.q.IndexSubject.IndexID.Eq(id), r.q.Subject.Nsfw.Is(true)).Count()
	if err != nil {
		r.log.Error("unexpected error when checking index nsfw", zap.Uint32("index_id", id))
		return false, errgo.Wrap(err, "dal")
	}

	return i != 0, nil
}

func (r mysqlRepo) Get(ctx context.Context, id model.IndexID) (model.Index, error) {
	i, err := r.q.Index.WithContext(ctx).
		Where(r.q.Index.ID.Eq(id), r.q.Index.Ban.Is(false)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Index{}, gerr.ErrNotFound
		}

		return model.Index{}, errgo.Wrap(err, "dal")
	}

	nsfw, err := r.isNsfw(ctx, id)
	if err != nil {
		return model.Index{}, err
	}

	ret := daoToModel(i)
	ret.NSFW = nsfw
	return *ret, nil
}

func (r mysqlRepo) New(ctx context.Context, i *model.Index) error {
	dao := modelToDAO(i)
	if err := r.q.Index.WithContext(ctx).Create(dao); err != nil {
		return errgo.Wrap(err, "failed to create index in db")
	}
	i.ID = dao.ID
	return nil
}

func (r mysqlRepo) Update(ctx context.Context, id model.IndexID, title string, desc string) error {
	query := r.q.Index.WithContext(ctx)
	result, err := query.Where(r.q.Index.ID.Eq(id)).Updates(dao.Index{
		Title: title,
		Desc:  desc,
	})
	return r.WrapResult(result, err, "failed to update index info")
}

func (r mysqlRepo) Delete(ctx context.Context, id model.IndexID) error {
	return r.q.Transaction(func(tx *query.Query) error {
		result, err := tx.Index.WithContext(ctx).Where(tx.Index.ID.Eq(id)).Delete()
		if err = r.WrapResult(result, err, "failed to delete index"); err != nil {
			return err
		}
		result, err = tx.IndexSubject.WithContext(ctx).Where(tx.IndexSubject.IndexID.Eq(id)).Delete()
		return r.WrapResult(result, err, "failed to delete subjects in the index")
	})
}

func (r mysqlRepo) CountSubjects(
	ctx context.Context, id model.IndexID, subjectType model.SubjectType,
) (int64, error) {
	q := r.q.IndexSubject.WithContext(ctx).Where(r.q.IndexSubject.IndexID.Eq(id))
	if subjectType != 0 {
		q = q.Where(r.q.IndexSubject.SubjectType.Eq(subjectType))
	}

	i, err := q.Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}

	return i, nil
}

func (r mysqlRepo) ListSubjects(
	ctx context.Context,
	id model.IndexID,
	subjectType model.SubjectType,
	limit, offset int,
) ([]Subject, error) {
	q := r.q.IndexSubject.WithContext(ctx).Joins(r.q.IndexSubject.Subject).
		Preload(r.q.IndexSubject.Subject.Fields).
		Where(r.q.IndexSubject.IndexID.Eq(id)).
		Order(r.q.IndexSubject.Order).
		Limit(limit).Offset(offset)
	if subjectType != 0 {
		q = q.Where(r.q.IndexSubject.SubjectType.Eq(subjectType))
	}

	d, err := q.Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make([]Subject, len(d))
	for i, s := range d {
		sub, err := subject.ConvertDao(&s.Subject)
		if err != nil {
			return nil, errgo.Wrap(err, "subject.ConvertDao")
		}

		results[i] = Subject{
			AddedAt: time.Unix(int64(s.CreatedTime), 0),
			Comment: s.Comment,
			Subject: sub,
		}
	}

	return results, nil
}

func (r mysqlRepo) AddOrUpdateIndexSubject(
	ctx context.Context, id model.IndexID,
	subjectID model.SubjectID, sort uint32, comment string,
) (*Subject, error) {
	index, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	subjectDO, err := r.q.Subject.WithContext(ctx).
		Where(r.q.Subject.ID.Eq(subjectID)).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gerr.ErrSubjectNotFound
		}
		return nil, errgo.Wrap(err, "dal")
	}

	subject, err := subject.ConvertDao(subjectDO)
	if err != nil {
		return nil, errgo.Wrap(err, "subject.ConvertDao")
	}

	indexSubject, err := r.q.IndexSubject.WithContext(ctx).
		Where(r.q.IndexSubject.IndexID.Eq(id), r.q.IndexSubject.SubjectID.Eq(subjectID)).
		FirstOrInit()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errgo.Wrap(err, "dal")
	}

	now := time.Now()

	if indexSubject.ID != 0 {
		// 已经存在，更新！
		err = r.updateIndexSubject(ctx, id, subjectID, sort, comment)
	} else {
		err = r.addSubjectToIndex(ctx, index, &dao.IndexSubject{
			Comment:     comment,
			Order:       sort,
			CreatedTime: uint32(now.Unix()),
			IndexID:     id,
			SubjectType: subjectDO.TypeID,
			SubjectID:   subjectID,
		})
	}

	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	return &Subject{
		AddedAt: now,
		Comment: comment,
		Subject: subject,
	}, nil
}

func (r mysqlRepo) updateIndexSubject(
	ctx context.Context, id model.IndexID,
	subjectID model.SubjectID, sort uint32, comment string,
) error {
	result, err := r.q.IndexSubject.WithContext(ctx).
		Where(r.q.IndexSubject.IndexID.Eq(id), r.q.IndexSubject.SubjectID.Eq(subjectID)).
		Updates(dao.IndexSubject{
			Order:   sort,
			Comment: comment,
		})
	return r.WrapResult(result, err, "failed to update index subject")
}

func (r mysqlRepo) addSubjectToIndex(ctx context.Context, index model.Index, subject *dao.IndexSubject) error {
	return r.q.Transaction(func(tx *query.Query) error {
		err := r.q.IndexSubject.WithContext(ctx).Create(subject)
		if err != nil {
			return errgo.Wrap(err, "failed to create subject in index")
		}

		result, err := r.q.Index.WithContext(ctx).
			Where(r.q.Index.ID.Eq(index.ID)).
			Updates(dao.Index{
				UpdatedTime:  uint32(time.Now().Unix()),
				SubjectCount: index.Total + 1,
			})

		return r.WrapResult(result, err, "failed to update index info")
	})
}

func (r mysqlRepo) DeleteIndexSubject(
	ctx context.Context, id model.IndexID, subjectID model.SubjectID,
) error {
	return r.q.Transaction(func(tx *query.Query) error {
		index, err := r.Get(ctx, id)
		if err != nil {
			return err
		}
		result, err := r.q.IndexSubject.WithContext(ctx).
			Where(r.q.IndexSubject.IndexID.Eq(id), r.q.IndexSubject.SubjectID.Eq(subjectID)).
			Delete()
		if err = r.WrapResult(result, err, "failed to delete index subject"); err != nil {
			return err
		}
		result, err = r.q.Index.WithContext(ctx).Where(r.q.Index.ID.Eq(id)).Updates(dao.Index{
			SubjectCount: index.Total - 1,
		})
		return r.WrapResult(result, err, "failed to update index info")
	})
}

func (r mysqlRepo) WrapResult(result gen.ResultInfo, err error, msg string) error {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gerr.ErrNotFound
		}
		return errgo.Wrap(err, msg)
	}
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func daoToModel(index *dao.Index) *model.Index {
	return &model.Index{
		ID:          index.ID,
		Title:       index.Title,
		Description: index.Desc,
		CreatorID:   index.CreatorID,
		Total:       index.SubjectCount,
		Comments:    index.ReplyCount,
		Collects:    index.CollectCount,
		Ban:         index.Ban,
		NSFW:        false, // check nsfw outSubjectIDe of this function
		CreatedAt:   time.Unix(int64(index.CreatedTime), 0),
		UpdatedAt:   time.Unix(int64(index.UpdatedTime), 0),
	}
}

func modelToDAO(index *model.Index) *dao.Index {
	return &dao.Index{
		ID:          index.ID,
		Type:        0,
		Title:       index.Title,
		Desc:        index.Description,
		CreatorID:   index.CreatorID,
		Ban:         index.Ban,
		CreatedTime: int32(index.CreatedAt.Unix()),
		UpdatedTime: uint32(index.UpdatedAt.Unix()),
	}
}
