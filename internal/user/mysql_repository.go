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

package user

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/strutil"
)

func NewUserRepo(q *query.Query, log *zap.Logger) (domain.UserRepo, error) {
	return mysqlRepo{q: q, log: log.Named("user.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) CountCollections(
	ctx context.Context,
	userID model.UserID,
	subjectType model.SubjectType,
	collectionType uint8,
	showPrivate bool,
) (int64, error) {
	q := m.q.SubjectCollection.WithContext(ctx).
		Where(m.q.SubjectCollection.UserID.Eq(userID))

	if subjectType != 0 {
		q = q.Where(m.q.SubjectCollection.SubjectType.Eq(subjectType))
	}

	if collectionType != 0 {
		q = q.Where(m.q.SubjectCollection.Type.Eq(collectionType))
	}

	if !showPrivate {
		q = q.Where(m.q.SubjectCollection.Private.Eq(model.CollectPrivacyNone))
	}

	c, err := q.Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}

	return c, nil
}

func (m mysqlRepo) ListCollections(
	ctx context.Context,
	userID model.UserID,
	subjectType model.SubjectType,
	collectionType uint8,
	showPrivate bool,
	limit, offset int,
) ([]model.Collection, error) {
	q := m.q.SubjectCollection.WithContext(ctx).
		Order(m.q.SubjectCollection.Lasttouch.Desc()).
		Where(m.q.SubjectCollection.UserID.Eq(userID)).Limit(limit).Offset(offset)

	if subjectType != 0 {
		q = q.Where(m.q.SubjectCollection.SubjectType.Eq(subjectType))
	}

	if collectionType != 0 {
		q = q.Where(m.q.SubjectCollection.Type.Eq(collectionType))
	}

	if !showPrivate {
		q = q.Where(m.q.SubjectCollection.Private.Eq(model.CollectPrivacyNone))
	}

	collections, err := q.Find()
	if err != nil {
		m.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make([]model.Collection, len(collections))
	for i, c := range collections {
		results[i] = model.Collection{
			UpdatedAt:   time.Unix(int64(c.Lasttouch), 0),
			Comment:     c.Comment,
			Tags:        strutil.Split(c.Tag, " "),
			SubjectType: c.SubjectType,
			Rate:        c.Rate,
			SubjectID:   c.SubjectID,
			EpStatus:    c.EpStatus,
			VolStatus:   c.VolStatus,
			Type:        c.Type,
			Private:     c.Private != model.CollectPrivacyNone,
		}
	}

	return results, nil
}

func (m mysqlRepo) GetCollection(
	ctx context.Context, userID model.UserID, subjectID model.SubjectID,
) (model.Collection, error) {
	c, err := m.q.SubjectCollection.WithContext(ctx).
		Where(m.q.SubjectCollection.UserID.Eq(userID), m.q.SubjectCollection.SubjectID.Eq(subjectID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Collection{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err), log.UserID(userID), log.SubjectID(subjectID))
		return model.Collection{}, errgo.Wrap(err, "dal")
	}

	return model.Collection{
		UpdatedAt:   time.Unix(int64(c.Lasttouch), 0),
		Comment:     c.Comment,
		Tags:        strutil.Split(c.Tag, " "),
		SubjectType: c.SubjectType,
		Rate:        c.Rate,
		SubjectID:   c.SubjectID,
		EpStatus:    c.EpStatus,
		VolStatus:   c.VolStatus,
		Type:        c.Type,
		Private:     c.Private != model.CollectPrivacyNone,
	}, nil
}

func (m mysqlRepo) GetByID(ctx context.Context, userID model.UserID) (model.User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.ID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))
		return model.User{}, errgo.Wrap(err, "dal")
	}

	return fromDao(u), nil
}

func (m mysqlRepo) GetByName(ctx context.Context, username string) (model.User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.Username.Eq(username)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))
		return model.User{}, errgo.Wrap(err, "dal")
	}

	return fromDao(u), nil
}

func (m mysqlRepo) GetByIDs(ctx context.Context, ids ...model.UserID) (map[model.UserID]model.User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.ID.In(ids...)).Find()
	if err != nil {
		m.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var r = make(map[model.UserID]model.User, len(ids))

	for _, member := range u {
		r[member.ID] = fromDao(member)
	}

	return r, nil
}

func fromDao(m *dao.Member) model.User {
	return model.User{
		UserName:         m.Username,
		NickName:         m.Nickname,
		UserGroup:        m.Groupid,
		Avatar:           m.Avatar,
		Sign:             m.Sign,
		ID:               m.ID,
		RegistrationTime: time.Unix(m.Regdate, 0),
	}
}
