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
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/elliotchance/phpserialize"
	ms "github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/strutil"
)

var _ domain.CollectionRepo = mysqlRepo{}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.CollectionRepo, error) {
	columns, err := getAllNonIndexFields(dao.SubjectCollection{})
	if err != nil {
		return nil, err
	}

	return mysqlRepo{
		q:             q,
		log:           log.Named("collection.mysqlRepo"),
		subjectUpsert: clause.OnConflict{DoUpdates: clause.AssignmentColumns(columns)},
	}, nil
}

type mysqlRepo struct {
	q             *query.Query
	log           *zap.Logger
	subjectUpsert clause.OnConflict
}

func (r mysqlRepo) UpdateSubjectCollection(
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
		VolStatus:   data.VolStatus,
		UpdatedAt:   uint32(data.UpdatedAt.Unix()),
		Private:     0,
	}

	switch data.Type {
	case model.CollectionTypeWish:
		d.WishAt = d.UpdatedAt
	case model.CollectionTypeDone:
		d.DoneAt = d.UpdatedAt
	case model.CollectionTypeDoing:
		d.DoingAt = d.UpdatedAt
	case model.CollectionTypeOnHold:
		d.OnHoldAt = d.UpdatedAt
	case model.CollectionTypeDropped:
		d.DroppedAt = d.UpdatedAt
	case model.CollectionTypeAll:
		// do nothing
	}

	err := r.q.SubjectCollection.WithContext(ctx).Clauses(r.subjectUpsert).Create(d)
	if err != nil {
		r.log.Error("unexpected error happened when updating subject collection", zap.Error(err),
			log.UserID(userID), log.SubjectID(subjectID), zap.Reflect("dao", d), zap.Reflect("data", data))
		return errgo.Wrap(err, "dal")
	}

	return nil
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

	if subjectType != 0 {
		q = q.Where(r.q.SubjectCollection.SubjectType.Eq(subjectType))
	}

	if collectionType != 0 {
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

	if subjectType != 0 {
		q = q.Where(r.q.SubjectCollection.SubjectType.Eq(subjectType))
	}

	if collectionType != 0 {
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

type epCollectionItem struct {
	EpisodeID model.EpisodeID      `ms:"eid"`
	Type      model.CollectionType `ms:"type"`
}

var errEpisodeInvalid = errors.New("number is not valid as episode ID")

type epCollection = map[model.EpisodeID]epCollectionItem

func deserializePhpEpStatus(phpSerialized []byte) (epCollection, error) {
	var e map[interface{}]interface{}
	if err := phpserialize.Unmarshal(phpSerialized, &e); err != nil {
		return nil, errgo.Wrap(err, "php deserialize")
	}

	var ep = make(epCollection, len(e))
	for key, value := range e {
		iKey, ok := key.(int64)
		if !ok {
			return nil, fmt.Errorf("failed to convert type %s to int64, value %v", reflect.TypeOf(key).String(), key)
		}
		if iKey <= 0 || iKey > math.MaxUint32 {
			return nil, errgo.Wrap(errEpisodeInvalid, strconv.FormatInt(iKey, 10))
		}

		var e epCollectionItem
		decoder, err := ms.NewDecoder(&ms.DecoderConfig{
			ErrorUnused:          true,
			ErrorUnset:           true,
			WeaklyTypedInput:     true,
			Result:               &e,
			TagName:              "ms",
			IgnoreUntaggedFields: true,
		})
		if err != nil {
			return nil, errgo.Wrap(err, "mapstructure.MewDecoder")
		}

		if err = decoder.Decode(value); err != nil {
			return nil, errgo.Wrap(err, "mapstructure.Decode")
		}

		ep[model.EpisodeID(iKey)] = e
	}

	return ep, nil
}
