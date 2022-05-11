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

package user

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/strutil"
	"github.com/bangumi/server/model"
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
	userID uint32,
	subjectType model.SubjectType,
	collectionType uint8,
	showPrivate bool,
) (int64, error) {
	q := m.q.SubjectCollection.WithContext(ctx).
		Where(m.q.SubjectCollection.UID.Eq(userID))

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
	userID uint32,
	subjectType model.SubjectType,
	collectionType uint8,
	showPrivate bool,
	limit, offset int,
) ([]model.Collection, error) {
	q := m.q.SubjectCollection.WithContext(ctx).
		Order(m.q.SubjectCollection.Lasttouch.Desc()).
		Where(m.q.SubjectCollection.UID.Eq(userID)).Limit(limit).Offset(offset)

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

func (m mysqlRepo) GetByID(ctx context.Context, userID uint32) (model.User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.UID.Eq(userID)).First()
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

func (m mysqlRepo) GetByIDs(ctx context.Context, ids ...uint32) (map[uint32]model.User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.UID.In(ids...)).Find()
	if err != nil {
		m.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var r = make(map[uint32]model.User, len(ids))

	for _, member := range u {
		r[member.UID] = fromDao(member)
	}

	return r, nil
}

func fromDao(m *dao.Member) model.User {
	return model.User{
		UserName:  m.Username,
		NickName:  m.Nickname,
		UserGroup: m.Groupid,
		Avatar:    m.Avatar,
		Sign:      m.Sign,
		ID:        m.UID,
	}
}
