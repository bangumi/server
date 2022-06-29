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
	"time"

	"go.uber.org/zap"
	"gorm.io/gen"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/strutil"
)

var _ domain.CollectionRepo = mysqlRepo{}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.CollectionRepo, error) {
	return mysqlRepo{
		q:   q,
		log: log.Named("collection.mysqlRepo"),
	}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (r mysqlRepo) CountSubjectCollections(
	ctx context.Context,
	userID model.UserID,
	subjectType model.SubjectType,
	collectionType model.SubjectCollectionType,
	showPrivate bool,
) (int64, error) {
	q := r.q.SubjectCollection.WithContext(ctx).
		Where(r.q.SubjectCollection.UserID.Eq(userID))

	if subjectType != model.SubjectTypeAll {
		q = q.Where(r.q.SubjectCollection.SubjectType.Eq(subjectType))
	}

	if collectionType != model.SubjectCollectionAll {
		q = q.Where(r.q.SubjectCollection.Type.Eq(uint8(collectionType)))
	}

	if !showPrivate {
		q = q.Where(r.q.SubjectCollection.Private.Eq(model.CollectPrivacyNone))
	}

	c, err := q.Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}

	return c, nil
}

func (r mysqlRepo) ListSubjectCollection(
	ctx context.Context,
	userID model.UserID,
	subjectType model.SubjectType,
	collectionType model.SubjectCollectionType,
	showPrivate bool,
	limit, offset int,
) ([]model.SubjectCollection, error) {
	q := r.q.SubjectCollection.WithContext(ctx).
		Order(r.q.SubjectCollection.UpdatedTime.Desc()).
		Where(r.q.SubjectCollection.UserID.Eq(userID)).Limit(limit).Offset(offset)

	if subjectType != model.SubjectTypeAll {
		q = q.Where(r.q.SubjectCollection.SubjectType.Eq(subjectType))
	}

	if collectionType != model.SubjectCollectionAll {
		q = q.Where(r.q.SubjectCollection.Type.Eq(uint8(collectionType)))
	}

	if !showPrivate {
		q = q.Where(r.q.SubjectCollection.Private.Eq(model.CollectPrivacyNone))
	}

	collections, err := q.Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make([]model.SubjectCollection, len(collections))
	for i, c := range collections {
		results[i] = model.SubjectCollection{
			UpdatedAt:   time.Unix(int64(c.UpdatedTime), 0),
			Comment:     c.Comment,
			Tags:        strutil.Split(c.Tag, " "),
			SubjectType: c.SubjectType,
			Rate:        c.Rate,
			SubjectID:   c.SubjectID,
			EpStatus:    c.EpStatus,
			VolStatus:   c.VolStatus,
			Type:        model.SubjectCollectionType(c.Type),
			Private:     c.Private != model.CollectPrivacyNone,
		}
	}

	return results, nil
}

func (r mysqlRepo) GetSubjectCollection(
	ctx context.Context, userID model.UserID, subjectID model.SubjectID,
) (model.SubjectCollection, error) {
	c, err := r.q.SubjectCollection.WithContext(ctx).
		Where(r.q.SubjectCollection.UserID.Eq(userID), r.q.SubjectCollection.SubjectID.Eq(subjectID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.SubjectCollection{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err), log.UserID(userID), log.SubjectID(subjectID))
		return model.SubjectCollection{}, errgo.Wrap(err, "dal")
	}

	return model.SubjectCollection{
		UpdatedAt:   time.Unix(int64(c.UpdatedTime), 0),
		Comment:     c.Comment,
		Tags:        strutil.Split(c.Tag, " "),
		SubjectType: c.SubjectType,
		Rate:        c.Rate,
		SubjectID:   c.SubjectID,
		EpStatus:    c.EpStatus,
		VolStatus:   c.VolStatus,
		Type:        model.SubjectCollectionType(c.Type),
		Private:     c.Private != model.CollectPrivacyNone,
	}, nil
}

func (r mysqlRepo) createEpisodeCollection(
	ctx context.Context,
	tx *query.Query,
	userID model.UserID,
	subjectID model.SubjectID,
	episodeID model.EpisodeID,
	collectionType model.EpisodeCollectionType,
	updatedAt time.Time,
) error {
	var e = make(mysqlEpCollection, 1)

	e[episodeID] = mysqlEpCollectionItem{EpisodeID: episodeID, Type: collectionType}

	bytes, err := serializePhpEpStatus(e)
	if err != nil {
		return err
	}

	err = tx.EpCollection.WithContext(ctx).
		Create(&dao.EpCollection{
			UserID:      userID,
			SubjectID:   subjectID,
			Status:      bytes,
			UpdatedTime: uint32(updatedAt.Unix()),
		})
	if err != nil {
		return errgo.Wrap(err, "gorm.Create")
	}

	return nil
}

func (r mysqlRepo) UpdateEpisodeCollection(
	ctx context.Context,
	userID model.UserID,
	subjectID model.SubjectID,
	episodeID model.EpisodeID,
	collectionType model.EpisodeCollectionType,
	updatedAt time.Time,
) error {
	updateTime := uint32(updatedAt.Unix())
	where := []gen.Condition{r.q.EpCollection.UserID.Eq(userID), r.q.EpCollection.SubjectID.Eq(subjectID)}
	return r.q.Transaction(func(tx *query.Query) error {
		d, err := tx.EpCollection.WithContext(ctx).Where(where...).First()
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				r.log.Error("failed to get episode collection record",
					zap.Error(err), log.UserID(userID), log.SubjectID(subjectID))
				return errgo.Wrap(err, "dal")
			}

			err = r.createEpisodeCollection(ctx, tx, userID, subjectID, episodeID, collectionType, updatedAt)
			r.log.Error("failed to create episode collection")
			return errgo.Wrap(err, "r.createEpisodeCollection")
		}

		var e mysqlEpCollection
		e, err = deserializePhpEpStatus(d.Status)
		if err != nil {
			r.log.Error("failed to deserialize php-serialized bytes to go data",
				zap.Error(err), log.UserID(userID), log.SubjectID(subjectID))
			return err
		}

		if v, ok := e[episodeID]; ok && v.Type == collectionType {
			return nil
		}

		e[episodeID] = mysqlEpCollectionItem{EpisodeID: episodeID, Type: collectionType}
		bytes, err := serializePhpEpStatus(e)
		if err != nil {
			return err
		}

		_, err = tx.SubjectCollection.WithContext(ctx).
			Where(tx.SubjectCollection.UserID.Eq(userID), tx.SubjectCollection.SubjectID.Eq(subjectID)).UpdateColumnSimple(
			tx.SubjectCollection.EpStatus.Value(countWatchedEp(e)), tx.SubjectCollection.UpdatedTime.Value(updateTime),
		)
		if err != nil {
			return errgo.Wrap(err, "SubjectCollection.UpdateSimple")
		}

		_, err = tx.EpCollection.WithContext(ctx).Where(where...).
			UpdateColumnSimple(r.q.EpCollection.Status.Value(bytes), r.q.EpCollection.UpdatedTime.Value(updateTime))
		if err != nil {
			return errgo.Wrap(err, "EpCollection.UpdateColumnSimple")
		}

		return nil
	})
}

func countWatchedEp(m mysqlEpCollection) uint32 {
	var count uint32
	for _, item := range m {
		if item.Type == model.EpisodeCollectionDone {
			count++
		}
	}

	return count
}

func (r mysqlRepo) GetEpisodeCollection(
	ctx context.Context, userID model.UserID, subjectID model.SubjectID,
) (model.EpisodeCollection, error) {
	d, err := r.q.EpCollection.WithContext(ctx).
		Where(r.q.EpCollection.UserID.Eq(userID), r.q.EpCollection.SubjectID.Eq(subjectID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		r.log.Error("failed to get episode collection record", zap.Error(err), log.UserID(userID), log.SubjectID(subjectID))
		return nil, errgo.Wrap(err, "dal")
	}

	e, err := deserializePhpEpStatus(d.Status)
	if err != nil {
		r.log.Error("failed to deserialize php-serialized bytes to go data",
			zap.Error(err), log.UserID(userID), log.SubjectID(subjectID))
		return nil, err
	}

	var result = make(model.EpisodeCollection, len(e))
	for id, item := range e {
		result[id] = model.EpisodeCollectionItem{
			ID:   item.EpisodeID,
			Type: item.Type,
		}
	}

	return result, nil
}
