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

package collection

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/strutil"
)

func NewCollectionRepo(q *query.Query, log *zap.Logger) (domain.CollectionRepo, error) {
	return mysqlRepo{q: q, log: log.Named("collection.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) UpdateCollection(
	ctx context.Context, userID model.UserID, subjectID model.SubjectID, data model.SubjectCollectionUpdate,
) error {

	var d = &dao.SubjectCollection{
		UserID:      userID,
		SubjectID:   subjectID,
		SubjectType: data.SubjectType,
		Rate:        data.Rate,
		Type:        uint8(data.Type),
		HasComment:  data.Comment != "",
		Comment:     data.Comment,
		Tag:         strings.Join(data.Tags, " "),
		EpStatus:    data.EpStatus,
		VolStatus:   data.EpStatus,
		UpdatedAt:   uint32(data.UpdatedAt.Unix()),
		Private:     0,
	}

	switch data.Type {
	case model.CollectionWish:
		d.WishAt = d.UpdatedAt
	case model.CollectionDone:
		d.DoneAt = d.UpdatedAt
	case model.CollectionDoing:
		d.DoingAt = d.UpdatedAt
	case model.CollectionOnHold:
		d.OnHoldAt = d.UpdatedAt
	case model.CollectionDropped:
		d.DroppedAt = d.UpdatedAt
	}

	err := m.q.SubjectCollection.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "interest_uid"}, {Name: "interest_subject_id"}},
		UpdateAll: true,
	}).Create(d)

	if err != nil {
		m.log.Error("unexpected error happened when updating subject collection", zap.Error(err),
			log.UserID(userID), log.SubjectID(subjectID), zap.Reflect("dao", d), zap.Reflect("data", data))
		return errgo.Wrap(err, "dal")
	}

	return nil
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
) ([]model.SubjectCollection, error) {
	q := m.q.SubjectCollection.WithContext(ctx).
		Order(m.q.SubjectCollection.UpdatedAt.Desc()).
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

	var results = make([]model.SubjectCollection, len(collections))
	for i, c := range collections {
		results[i] = model.SubjectCollection{
			UpdatedAt:   time.Unix(int64(c.UpdatedAt), 0),
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
) (model.SubjectCollection, error) {
	c, err := m.q.SubjectCollection.WithContext(ctx).
		Where(m.q.SubjectCollection.UserID.Eq(userID), m.q.SubjectCollection.SubjectID.Eq(subjectID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.SubjectCollection{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err), log.UserID(userID), log.SubjectID(subjectID))
		return model.SubjectCollection{}, errgo.Wrap(err, "dal")
	}

	return model.SubjectCollection{
		UpdatedAt:   time.Unix(int64(c.UpdatedAt), 0),
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
