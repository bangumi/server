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
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/strutil"
	"github.com/bangumi/server/internal/pkg/timex"
)

var _ domain.CollectionRepo = mysqlRepo{}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.CollectionRepo, error) {
	columns, err := getAllNonIndexFields(dao.SubjectCollection{})
	if err != nil {
		return nil, err
	}

	return mysqlRepo{
		q:   q,
		log: log.Named("collection.mysqlRepo"),

		subjectUpsert: clause.OnConflict{DoUpdates: clause.AssignmentColumns(columns)},
	}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger

	subjectUpsert clause.OnConflict
}

func (r mysqlRepo) UpdateSubjectCollection(
	ctx context.Context, userID model.UserID, subjectID model.SubjectID, data model.SubjectCollectionUpdate,
) error {
	if data.Type == model.CollectionTypeAll {
		return fmt.Errorf("%w: can't set collection type to 0", domain.ErrInvalidInput)
	}

	where := []gen.Condition{r.q.SubjectCollection.SubjectID.Eq(subjectID), r.q.SubjectCollection.UserID.Eq(userID)}

	return r.q.Transaction(func(tx *query.Query) error {
		old, err := tx.SubjectCollection.WithContext(ctx).Where(where...).First()
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return fmt.Errorf(
					"%w: subject collection not found, user should add it to collection from WEB", domain.ErrInvalidInput)
			}

			return errgo.Wrap(err, "dal")
		}

		var d = &dao.SubjectCollection{
			UserID:      userID,
			SubjectID:   subjectID,
			Rate:        data.Rate,
			Type:        uint8(data.Type),
			HasComment:  data.Comment != "",
			Comment:     data.Comment,
			Tag:         strings.Join(data.Tags, " "),
			EpStatus:    data.EpStatus,
			VolStatus:   data.VolStatus,
			SubjectType: old.SubjectType,
			WishAt:      old.WishAt,
			DoingAt:     old.DoingAt,
			DoneAt:      old.DoneAt,
			OnHoldAt:    old.OnHoldAt,
			DroppedAt:   old.DroppedAt,
			UpdatedAt:   uint32(data.UpdatedAt.Unix()),
			Private:     old.Private,
		}

		if old.Private != model.CollectPrivacyBan {
			updatePrivate(d, data.Private)
		}

		updateTimeStamp(d, model.CollectionType(old.Type), data.Type, uint32(data.UpdatedAt.Unix()))

		_, err = tx.SubjectCollection.WithContext(ctx).Debug().Omit(
			r.q.SubjectCollection.ID, r.q.SubjectCollection.UserID,
			r.q.SubjectCollection.SubjectID, r.q.SubjectCollection.SubjectType,
		).Where(where...).UpdateColumns(d)
		if err != nil {
			r.log.Error("unexpected error happened when updating subject collection", zap.Error(err),
				log.UserID(userID), log.SubjectID(subjectID), zap.Reflect("dao", d), zap.Reflect("data", data))
			return errgo.Wrap(err, "dal")
		}

		return nil
	})
}

func updatePrivate(newRecord *dao.SubjectCollection, private bool) {
	if private {
		newRecord.Private = model.CollectPrivacySelf
	} else {
		newRecord.Private = model.CollectPrivacyNone
	}
}

// update new record timestamp base on new Type.
func updateTimeStamp(newRecord *dao.SubjectCollection, oldType, newType model.CollectionType, updatedAt uint32) {
	if oldType == newType {
		return
	}

	switch newType {
	case model.CollectionTypeWish:
		newRecord.WishAt = updatedAt
	case model.CollectionTypeDone:
		newRecord.DoneAt = updatedAt
	case model.CollectionTypeDoing:
		newRecord.DoingAt = updatedAt
	case model.CollectionTypeOnHold:
		newRecord.OnHoldAt = updatedAt
	case model.CollectionTypeDropped:
		newRecord.DroppedAt = updatedAt
	case model.CollectionTypeAll:
		// already checked, do nothing
	}
}

func (r mysqlRepo) CountSubjectCollections(
	ctx context.Context,
	userID model.UserID,
	subjectType model.SubjectType,
	collectionType model.CollectionType,
	showPrivate bool,
) (int64, error) {
	q := r.q.SubjectCollection.WithContext(ctx).
		Where(r.q.SubjectCollection.UserID.Eq(userID))

	if subjectType != model.SubjectTypeAll {
		q = q.Where(r.q.SubjectCollection.SubjectType.Eq(subjectType))
	}

	if collectionType != model.CollectionTypeAll {
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
	collectionType model.CollectionType,
	showPrivate bool,
	limit, offset int,
) ([]model.SubjectCollection, error) {
	q := r.q.SubjectCollection.WithContext(ctx).
		Order(r.q.SubjectCollection.UpdatedAt.Desc()).
		Where(r.q.SubjectCollection.UserID.Eq(userID)).Limit(limit).Offset(offset)

	if subjectType != model.SubjectTypeAll {
		q = q.Where(r.q.SubjectCollection.SubjectType.Eq(subjectType))
	}

	if collectionType != model.CollectionTypeAll {
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
			UpdatedAt:   time.Unix(int64(c.UpdatedAt), 0),
			Comment:     c.Comment,
			Tags:        strutil.Split(c.Tag, " "),
			SubjectType: c.SubjectType,
			Rate:        c.Rate,
			SubjectID:   c.SubjectID,
			EpStatus:    c.EpStatus,
			VolStatus:   c.VolStatus,
			Type:        model.CollectionType(c.Type),
			Private:     c.Private != model.CollectPrivacyNone,
		}
	}

	return results, nil
}

func (r mysqlRepo) createEpisodeCollection(
	ctx context.Context,
	tx *query.Query,
	userID model.UserID,
	subjectID model.SubjectID,
	episodeID model.EpisodeID,
	collectionType model.CollectionType,
) error {
	var e = make(mysqlEpCollection, 1)

	e[episodeID] = mysqlEpCollectionItem{
		EpisodeID: episodeID,
		Type:      collectionType,
	}

	bytes, err := serializePhpEpStatus(e)
	if err != nil {
		return err
	}

	err = tx.EpCollection.WithContext(ctx).
		Create(&dao.EpCollection{
			UserID:    userID,
			SubjectID: subjectID,
			Status:    bytes,
			UpdatedAt: uint32(time.Now().Unix()),
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
	collectionType model.CollectionType,
) error {
	return r.q.Transaction(func(tx *query.Query) error {
		d, err := tx.EpCollection.WithContext(ctx).
			Where(r.q.EpCollection.UserID.Eq(userID), r.q.EpCollection.SubjectID.Eq(subjectID)).First()
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				r.log.Error("failed to get episode collection record",
					zap.Error(err), log.UserID(userID), log.SubjectID(subjectID))
				return errgo.Wrap(err, "dal")
			}

			err = r.createEpisodeCollection(ctx, tx, userID, subjectID, episodeID, collectionType)
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

		if v, ok := e[episodeID]; ok {
			if v.Type == collectionType {
				return nil
			}
		}

		e[episodeID] = mysqlEpCollectionItem{
			EpisodeID: episodeID,
			Type:      collectionType,
		}

		bytes, err := serializePhpEpStatus(e)
		if err != nil {
			return err
		}

		_, err = tx.EpCollection.WithContext(ctx).
			Where(r.q.EpCollection.UserID.Eq(userID), r.q.EpCollection.SubjectID.Eq(subjectID)).
			UpdateColumnSimple(r.q.EpCollection.Status.Value(bytes), r.q.EpCollection.UpdatedAt.Value(timex.NowU32()))
		if err != nil {
			return errgo.Wrap(err, "gorm.UpdateColumnSimple")
		}

		return nil
	})
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
		UpdatedAt:   time.Unix(int64(c.UpdatedAt), 0),
		Comment:     c.Comment,
		Tags:        strutil.Split(c.Tag, " "),
		SubjectType: c.SubjectType,
		Rate:        c.Rate,
		SubjectID:   c.SubjectID,
		EpStatus:    c.EpStatus,
		VolStatus:   c.VolStatus,
		Type:        model.CollectionType(c.Type),
		Private:     c.Private != model.CollectPrivacyNone,
	}, nil
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

func getAllNonIndexFields(v interface{}) ([]string, error) {
	rvType := reflect.TypeOf(v)
	s := make([]string, 0, rvType.NumField())
	for i := 0; i < rvType.NumField(); i++ {
		cfg := schema.ParseTagSetting(rvType.Field(i).Tag.Get("gorm"), ";")
		column := cfg[strings.ToUpper("column")]
		if column == "" {
			f := rvType.Field(i)
			//nolint:goerr113
			return nil, fmt.Errorf("failed to parse struct field %s, '%s' ,%v", f.Name, f.Tag.Get("gorm"), column)
		}
		if column == "interest_id" || column == "interest_uid" || column == "interest_subject_id" {
			continue
		}
		s = append(s, column)
	}

	return s, nil
}
