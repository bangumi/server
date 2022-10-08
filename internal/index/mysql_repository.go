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
	"gorm.io/gen"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/subject"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.IndexRepo, error) {
	return mysqlRepo{q: q, log: log.Named("index.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (r mysqlRepo) isNsfw(ctx context.Context, id model.IndexID) (bool, error) {
	i, err := r.q.IndexSubject.WithContext(ctx).
		Join(r.q.Subject, r.q.IndexSubject.Sid.EqCol(r.q.Subject.ID)).
		Where(r.q.IndexSubject.Rid.Eq(id), r.q.Subject.Nsfw.Is(true)).Count()
	if err != nil {
		r.log.Error("unexpected error when checking index nsfw", zap.Uint32("index_id", id))
		return false, errgo.Wrap(err, "dal")
	}

	return i != 0, nil
}

func (r mysqlRepo) Get(ctx context.Context, id model.IndexID) (model.Index, error) {
	i, err := r.q.Index.WithContext(ctx).
		Where(r.q.Index.ID.Eq(id), r.q.Index.Ban.Is(false)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Index{}, domain.ErrNotFound
		}

		return model.Index{}, errgo.Wrap(err, "dal")
	}

	nsfw, err := r.isNsfw(ctx, id)
	if err != nil {
		return model.Index{}, err
	}

	ret := indexDaoToModel(*i)
	ret.NSFW = nsfw
	return ret, nil
}

func (r mysqlRepo) New(ctx context.Context, i *model.Index) error {
	dao := indexModelToDAO(*i)
	if err := r.q.Index.WithContext(ctx).Create(&dao); err != nil {
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
		result, err = tx.IndexSubject.WithContext(ctx).Where(tx.IndexSubject.Rid.Eq(id)).Delete()
		if err = r.WrapResult(result, err, "failed to delete subjects in the index"); err != nil {
			return err
		}
		return nil
	})
}

func (r mysqlRepo) CountSubjects(
	ctx context.Context, id model.IndexID, subjectType model.SubjectType,
) (int64, error) {
	q := r.q.IndexSubject.WithContext(ctx).Where(r.q.IndexSubject.Rid.Eq(id))
	if subjectType != 0 {
		q = q.Where(r.q.IndexSubject.Type.Eq(subjectType))
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
) ([]domain.IndexSubject, error) {
	q := r.q.IndexSubject.WithContext(ctx).
		Joins(r.q.IndexSubject.Subject).
		Preload(r.q.IndexSubject.Subject.Fields).
		Where(r.q.IndexSubject.Rid.Eq(id)).
		Order(r.q.IndexSubject.Order).Limit(limit).Offset(offset)
	if subjectType != 0 {
		q = q.Where(r.q.IndexSubject.Type.Eq(subjectType))
	}

	d, err := q.Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make([]domain.IndexSubject, len(d))
	for i, s := range d {
		sub, err := subject.ConvertDao(&s.Subject)
		if err != nil {
			return nil, errgo.Wrap(err, "subject.ConvertDao")
		}

		results[i] = domain.IndexSubject{
			AddedAt: time.Unix(int64(s.Dateline), 0),
			Comment: s.Comment,
			Subject: sub,
		}
	}

	return results, nil
}

func (r mysqlRepo) AddIndexSubject(
	ctx context.Context,
	info domain.IndexEditSubjectInfo,
) (*domain.IndexSubject, error) {
	index, err := r.Get(ctx, info.IndexID)
	if err != nil {
		return nil, err
	}

	subjectDO, err := r.q.Subject.WithContext(ctx).
		Where(r.q.Subject.ID.Eq(info.SubjectID)).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSubjectNotFound
		}
		return nil, errgo.Wrap(err, "dal")
	}

	now := time.Now()

	indexSubject, err := r.q.IndexSubject.WithContext(ctx).
		Where(
			r.q.IndexSubject.Rid.Eq(info.IndexID),
			r.q.IndexSubject.Sid.Eq(uint32(info.SubjectID)),
		).
		First()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errgo.Wrap(err, "dal")
	}

	if indexSubject != nil && indexSubject.ID != 0 {
		return nil, domain.ErrExists
	}

	err = r.addSubjectToIndex(ctx, index, &dao.IndexSubject{
		Comment:  info.Comment,
		Order:    info.Sort,
		Dateline: uint32(now.Unix()),
		Type:     subjectDO.TypeID,
		Rid:      info.IndexID,
		Sid:      uint32(info.SubjectID),
	})

	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	subject, err := subject.ConvertDao(subjectDO)
	if err != nil {
		return nil, errgo.Wrap(err, "subject.ConvertDao")
	}

	return &domain.IndexSubject{
		AddedAt: now,
		Comment: info.Comment,
		Subject: subject,
	}, nil
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
				Lasttouch:    uint32(time.Now().Unix()),
				SubjectTotal: index.Total + 1,
			})

		return r.WrapResult(result, err, "failed to update index info")
	})
}

func (r mysqlRepo) UpdateIndexSubject(ctx context.Context, info domain.IndexEditSubjectInfo) error {
	q := r.q.IndexSubject
	result, err := q.WithContext(ctx).
		Where(q.Rid.Eq(info.IndexID), q.Sid.Eq(uint32(info.SubjectID))).
		Updates(dao.IndexSubject{
			Comment: info.Comment,
			Order:   info.Sort,
		})
	return r.WrapResult(result, err, "failed to update index subject")
}

func (r mysqlRepo) DeleteIndexSubject(ctx context.Context, info domain.IndexSubjectInfo) error {
	return r.q.Transaction(func(tx *query.Query) error {
		index, err := r.Get(ctx, info.IndexID)
		if err != nil {
			return err
		}
		result, err := r.q.IndexSubject.WithContext(ctx).
			Where(
				r.q.IndexSubject.Rid.Eq(info.IndexID),
				r.q.IndexSubject.Sid.Eq(uint32(info.SubjectID)),
			).Delete()
		if err = r.WrapResult(result, err, "failed to delete index subject"); err != nil {
			return err
		}
		result, err = r.q.Index.WithContext(ctx).
			Where(r.q.Index.ID.Eq(info.IndexID)).
			Updates(dao.Index{
				SubjectTotal: index.Total - 1,
			})
		if err = r.WrapResult(result, err, "failed to update index info"); err != nil {
			return err
		}
		return nil
	})
}

func (r mysqlRepo) WrapResult(result gen.ResultInfo, err error, msg string) error {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		return errgo.Wrap(err, msg)
	}
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// 获取用户 creatorID 创建的目录.
func (r mysqlRepo) GetIndicesByUser(
	ctx context.Context, creatorID model.UserID, limit int, offset int,
) ([]model.Index, error) {
	query := r.q.Index.WithContext(ctx).
		Where(r.q.Index.CreatorID.Eq(creatorID)).
		Order(r.q.Index.Dateline).Limit(limit).Offset(offset)
	arr, err := query.Find()
	if err != nil {
		return nil, errgo.Wrap(err, "failed to get index")
	}
	ret := make([]model.Index, len(arr))
	for i := range arr {
		ret[i] = indexDaoToModel(*arr[i])
	}
	return ret, nil
}

// 获取用户 creatorID 收藏的目录.
func (r mysqlRepo) GetCollectedIndicesByUser(
	ctx context.Context, creatorID model.UserID, limit int, offset int,
) ([]model.IndexCollect, error) {
	query := r.q.IndexCollect.WithContext(ctx).
		Debug().
		Preload(r.q.IndexCollect.Index).
		Preload(r.q.IndexCollect.Index.Creator).
		Where(r.q.IndexCollect.CreatorID.Eq(creatorID)).
		Order(r.q.IndexCollect.CreatedTime).
		Limit(limit).Offset(offset)
	arr, err := query.Find()
	if err != nil {
		return nil, errgo.Wrap(err, "failed to get index that collected by user")
	}
	ret := make([]model.IndexCollect, len(arr))
	for i := range arr {
		ret[i] = indexCollectDaoToModel(*arr[i])
	}
	return ret, nil
}

func (r mysqlRepo) CollectIndex(ctx context.Context, id model.IndexID, uid model.UserID) error {
	// 查询是否存在
	iCollect, err := r.q.IndexCollect.WithContext(ctx).
		Where(r.q.IndexCollect.IndexID.Eq(id), r.q.IndexCollect.CreatorID.Eq(uid)).
		First()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errgo.Wrap(err, "failed to find collect record")
	}

	if iCollect != nil && iCollect.ID != 0 {
		return domain.ErrExists
	}

	return r.q.Transaction(func(tx *query.Query) error {
		// Index 的 Collect + 1
		index, err := r.Get(ctx, id)
		if err != nil {
			return err
		}
		result, err := r.q.Index.WithContext(ctx).
			Where(r.q.Index.ID.Eq(id)).
			Updates(dao.Index{
				Collects: index.Collects + 1,
			})
		if err = r.WrapResult(result, err, "failed to increases collect count"); err != nil {
			return err
		}

		// 写 IndexCollect 记录
		err = r.q.IndexCollect.WithContext(ctx).
			Create(&dao.IndexCollect{
				IndexID:     id,
				CreatorID:   uid,
				CreatedTime: uint32(time.Now().Unix()),
			})
		return errgo.Wrap(err, "write index collect record failed")
	})
}

func (r mysqlRepo) DeCollectIndex(ctx context.Context, id model.IndexID, uid model.UserID) error {
	// 查询是否存在
	_, err := r.q.IndexCollect.WithContext(ctx).
		Where(r.q.IndexCollect.IndexID.Eq(id), r.q.IndexCollect.CreatorID.Eq(uid)).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		return errgo.Wrap(err, "failed to find collect record")
	}

	return r.q.Transaction(func(tx *query.Query) error {
		// Index 的 Collect - 1
		index, err := r.Get(ctx, id)
		if err != nil {
			return err
		}
		result, err := r.q.Index.WithContext(ctx).
			Where(r.q.Index.ID.Eq(id)).
			Updates(dao.Index{
				Collects: index.Collects - 1,
			})
		if err = r.WrapResult(result, err, "failed to increases collect count"); err != nil {
			return err
		}

		// 删除 IndexCollect 记录
		result, err = r.q.IndexCollect.WithContext(ctx).
			Where(r.q.IndexCollect.IndexID.Eq(id), r.q.IndexCollect.CreatorID.Eq(uid)).
			Delete()
		if err = r.WrapResult(result, err, "failed to increases collect count"); err != nil {
			return err
		}
		return nil
	})
}

func indexDaoToModel(index dao.Index) model.Index {
	return model.Index{
		ID:          index.ID,
		Title:       index.Title,
		Description: index.Desc,
		CreatorID:   index.CreatorID,
		Total:       index.SubjectTotal,
		Comments:    index.Replies,
		Collects:    index.Collects,
		Ban:         index.Ban,
		NSFW:        false, // check nsfw outside of this function
		CreatedAt:   time.Unix(int64(index.Dateline), 0),
		UpdateAt:    time.Unix(int64(index.Lasttouch), 0),
	}
}

func indexModelToDAO(index model.Index) dao.Index {
	return dao.Index{
		ID:        index.ID,
		Type:      0,
		Title:     index.Title,
		Desc:      index.Description,
		CreatorID: index.CreatorID,
		Ban:       index.Ban,
		Dateline:  int32(index.CreatedAt.Unix()),
		Lasttouch: uint32(index.UpdateAt.Unix()),
	}
}

func indexCollectDaoToModel(i dao.IndexCollect) model.IndexCollect {
	user := i.Index.Creator
	return model.IndexCollect{
		ID:        i.ID,
		CreatorID: i.CreatorID,
		CreatedAt: time.Unix(int64(i.CreatedTime), 0),
		Index:     indexDaoToModel(i.Index),
		IndexCreator: model.User{
			ID:               user.ID,
			NickName:         user.Nickname,
			UserName:         user.Username,
			Avatar:           user.Avatar,
			Sign:             user.Sign,
			RegistrationTime: time.Unix(user.Regdate, 0),
			UserGroup:        user.Groupid,
		},
	}
}
