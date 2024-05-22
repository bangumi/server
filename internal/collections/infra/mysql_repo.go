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

package infra

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/trim21/errgo"
	"github.com/trim21/go-phpserialize"
	"go.uber.org/zap"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/dal/utiltype"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/collections"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/internal/subject"
)

var _ collections.Repo = mysqlRepo{}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (collections.Repo, error) {
	return mysqlRepo{
		q:   q,
		log: log.Named("collection.mysqlRepo"),
	}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (r mysqlRepo) getSubjectCollection(
	ctx context.Context, user model.UserID, subject model.SubjectID,
) (*dao.SubjectCollection, error) {
	s, err := r.q.SubjectCollection.WithContext(ctx).
		Where(r.q.SubjectCollection.UserID.Eq(user), r.q.SubjectCollection.SubjectID.Eq(subject)).Take()
	if err != nil {
		return nil, gerr.WrapGormError(err)
	}

	return s, nil
}

func (r mysqlRepo) convertToSubjectCollection(s *dao.SubjectCollection) (*collection.Subject, error) {
	return collection.NewSubjectCollection(
		s.SubjectID,
		s.UserID,
		s.Rate,
		collection.SubjectCollection(s.Type),
		s.Comment.String(),
		collection.CollectPrivacy(s.Private),
		gstr.Split(s.Tag, " "),
		s.VolStatus,
		s.EpStatus,
	)
}

func (r mysqlRepo) UpdateSubjectCollection(
	ctx context.Context,
	userID model.UserID,
	subject model.Subject,
	at time.Time,
	ip string,
	update func(ctx context.Context, s *collection.Subject) (*collection.Subject, error),
) error {
	s, err := r.getSubjectCollection(ctx, userID, subject.ID)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return gerr.ErrSubjectNotCollected
		}
		return err
	}
	return r.updateOrCreateSubjectCollection(ctx, userID, subject, at, ip, update, s)
}

func (r mysqlRepo) UpdateOrCreateSubjectCollection(
	ctx context.Context,
	userID model.UserID,
	subject model.Subject,
	at time.Time,
	ip string,
	update func(ctx context.Context, s *collection.Subject) (*collection.Subject, error),
) error {
	s, err := r.getSubjectCollection(ctx, userID, subject.ID)
	if err != nil {
		if !errors.Is(err, gerr.ErrNotFound) {
			return err
		}
		s = nil
	}
	return r.updateOrCreateSubjectCollection(ctx, userID, subject, at, ip, update, s)
}

func (r mysqlRepo) updateOrCreateSubjectCollection(
	ctx context.Context,
	userID model.UserID,
	subject model.Subject,
	at time.Time,
	ip string,
	update func(ctx context.Context, s *collection.Subject) (*collection.Subject, error),
	obj *dao.SubjectCollection,
) error {
	created := obj == nil
	if created {
		obj = &dao.SubjectCollection{
			SubjectID:   subject.ID,
			SubjectType: subject.TypeID,
			UserID:      userID,
		}
	}
	collectionSubject, err := r.convertToSubjectCollection(obj)
	if err != nil {
		return errgo.Trace(err)
	}

	original := *collectionSubject

	// Update subject collection
	s, err := update(ctx, collectionSubject)
	if err != nil {
		return errgo.Trace(err)
	}

	if err = r.updateSubjectCollection(obj, &original, s, at, ip, created); err != nil {
		return errgo.Trace(err)
	}

	T := r.q.SubjectCollection
	if created {
		err = errgo.Trace(T.WithContext(ctx).Create(obj))
	} else {
		err = errgo.Trace(T.WithContext(ctx).Save(obj))
	}

	if err != nil {
		return err
	}

	r.updateSubject(ctx, subject.ID)
	return nil
}

// 根据新旧 collection.Subject 状态
// 更新 dao.SubjectCollection.
func (r mysqlRepo) updateSubjectCollection(
	obj *dao.SubjectCollection,
	original *collection.Subject,
	s *collection.Subject,
	at time.Time,
	ip string,
	isNew bool,
) error {
	obj.UpdatedTime = uint32(at.Unix())
	obj.Comment = utiltype.HTMLEscapedString(s.Comment())
	obj.HasComment = s.Comment() != ""
	obj.Tag = strings.Join(s.Tags(), " ")
	obj.EpStatus = s.Eps()
	obj.VolStatus = s.Vols()
	obj.Rate = s.Rate()
	obj.Private = uint8(s.Privacy())
	obj.Type = uint8(s.TypeID())

	if s.TypeID() != original.TypeID() {
		err := r.updateCollectionTime(obj, s.TypeID(), at)
		if err != nil {
			return errgo.Trace(err)
		}
	}

	// Update IP
	if ip != "" {
		obj.LastUpdateIP = ip
		if isNew {
			obj.CreateIP = ip
		}
	}
	return nil
}

func (r mysqlRepo) WithQuery(query *query.Query) collections.Repo {
	return mysqlRepo{q: query, log: r.log}
}

func (r mysqlRepo) CountSubjectCollections(
	ctx context.Context,
	userID model.UserID,
	subjectType model.SubjectType,
	collectionType collection.SubjectCollection,
	showPrivate bool,
) (int64, error) {
	q := r.q.SubjectCollection.WithContext(ctx).
		Where(r.q.SubjectCollection.UserID.Eq(userID))

	if subjectType != model.SubjectTypeAll {
		q = q.Where(r.q.SubjectCollection.SubjectType.Eq(subjectType))
	}

	if collectionType != collection.SubjectCollectionAll {
		q = q.Where(r.q.SubjectCollection.Type.Eq(uint8(collectionType)))
	}

	if !showPrivate {
		q = q.Where(r.q.SubjectCollection.Private.Eq(uint8(collection.CollectPrivacyNone)))
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
	collectionType collection.SubjectCollection,
	showPrivate bool,
	limit, offset int,
) ([]collection.UserSubjectCollection, error) {
	q := r.q.SubjectCollection.WithContext(ctx).
		Order(r.q.SubjectCollection.UpdatedTime.Desc()).
		Where(r.q.SubjectCollection.UserID.Eq(userID)).Limit(limit).Offset(offset)

	if subjectType != model.SubjectTypeAll {
		q = q.Where(r.q.SubjectCollection.SubjectType.Eq(subjectType))
	}

	if collectionType != collection.SubjectCollectionAll {
		q = q.Where(r.q.SubjectCollection.Type.Eq(uint8(collectionType)))
	}

	if !showPrivate {
		q = q.Where(r.q.SubjectCollection.Private.Eq(uint8(collection.CollectPrivacyNone)))
	}

	collections, err := q.Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make([]collection.UserSubjectCollection, len(collections))
	for i, c := range collections {
		results[i] = collection.UserSubjectCollection{
			UpdatedAt:   time.Unix(int64(c.UpdatedTime), 0),
			Comment:     string(c.Comment),
			Tags:        gstr.Split(c.Tag, " "),
			SubjectType: c.SubjectType,
			Rate:        c.Rate,
			SubjectID:   c.SubjectID,
			EpStatus:    c.EpStatus,
			VolStatus:   c.VolStatus,
			Type:        collection.SubjectCollection(c.Type),
			Private:     collection.CollectPrivacy(c.Private) != collection.CollectPrivacyNone,
		}
	}

	return results, nil
}

func (r mysqlRepo) GetSubjectCollection(
	ctx context.Context, userID model.UserID, subjectID model.SubjectID,
) (collection.UserSubjectCollection, error) {
	c, err := r.q.SubjectCollection.WithContext(ctx).
		Where(r.q.SubjectCollection.UserID.Eq(userID), r.q.SubjectCollection.SubjectID.Eq(subjectID)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return collection.UserSubjectCollection{}, gerr.ErrSubjectNotCollected
		}

		return collection.UserSubjectCollection{}, errgo.Wrap(err, "dal")
	}

	return collection.UserSubjectCollection{
		ID:          uint64(c.ID),
		UpdatedAt:   time.Unix(int64(c.UpdatedTime), 0),
		Comment:     string(c.Comment),
		Tags:        gstr.Split(c.Tag, " "),
		SubjectType: c.SubjectType,
		Rate:        c.Rate,
		SubjectID:   c.SubjectID,
		EpStatus:    c.EpStatus,
		VolStatus:   c.VolStatus,
		Type:        collection.SubjectCollection(c.Type),
		Private:     collection.CollectPrivacy(c.Private) != collection.CollectPrivacyNone,
	}, nil
}

func (r mysqlRepo) GetSubjectEpisodesCollection(
	ctx context.Context,
	userID model.UserID,
	subjectID model.SubjectID,
) (collection.UserSubjectEpisodesCollection, error) {
	d, err := r.q.EpCollection.WithContext(ctx).Where(
		r.q.EpCollection.UserID.Eq(userID),
		r.q.EpCollection.SubjectID.Eq(subjectID),
	).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return collection.UserSubjectEpisodesCollection{}, nil
		}
		return nil, errgo.Wrap(err, "query.EpCollection.Find")
	}

	e, err := deserializePhpEpStatus(d.Status)
	if err != nil {
		return nil, err
	}

	return e.toModel(), nil
}

func (r mysqlRepo) updateSubject(ctx context.Context, subjectID model.SubjectID) {
	if err := r.updateSubjectTags(ctx, subjectID); err != nil {
		r.log.Error("failed to update subject tags", zap.Error(err))
	}

	if err := r.reCountSubjectCollection(ctx, subjectID); err != nil {
		r.log.Error("failed to update collection counts", zap.Error(err))
	}
}

func (r mysqlRepo) reCountSubjectCollection(ctx context.Context, subjectID model.SubjectID) error {
	var counts []struct {
		Type  uint8  `gorm:"type"`
		Total uint32 `gorm:"total"`
	}

	err := r.q.SubjectCollection.WithContext(ctx).
		Select(r.q.SubjectCollection.Type.As("type"), r.q.SubjectCollection.Type.Count().As("total")).
		Group(r.q.SubjectCollection.Type).
		Where(r.q.SubjectCollection.SubjectID.Eq(subjectID)).Group(r.q.SubjectCollection.Type).Scan(&counts)
	if err != nil {
		return errgo.Wrap(err, "dal")
	}

	var updater = make([]field.AssignExpr, 0, 5)

	for _, count := range counts {
		switch collection.SubjectCollection(count.Type) { //nolint:exhaustive
		case collection.SubjectCollectionDropped:
			updater = append(updater, r.q.Subject.Dropped.Value(count.Total))

		case collection.SubjectCollectionWish:
			updater = append(updater, r.q.Subject.Wish.Value(count.Total))

		case collection.SubjectCollectionDoing:
			updater = append(updater, r.q.Subject.Doing.Value(count.Total))

		case collection.SubjectCollectionOnHold:
			updater = append(updater, r.q.Subject.OnHold.Value(count.Total))

		case collection.SubjectCollectionDone:
			updater = append(updater, r.q.Subject.Done.Value(count.Total))
		}
	}

	_, err = r.q.Subject.WithContext(ctx).Where(r.q.Subject.ID.Eq(subjectID)).UpdateSimple(updater...)
	if err != nil {
		return errgo.Wrap(err, "dal")
	}

	return nil
}

func (r mysqlRepo) updateSubjectTags(ctx context.Context, subjectID model.SubjectID) error {
	collections, err := r.q.SubjectCollection.WithContext(ctx).
		Where(
			r.q.SubjectCollection.SubjectID.Eq(subjectID),
			r.q.SubjectCollection.Private.Neq(uint8(collection.CollectPrivacyBan)),
		).Find()
	if err != nil {
		return errgo.Wrap(err, "failed to get all collection")
	}

	var tags = make(map[string]int)
	for _, collection := range collections {
		for _, s := range strings.Split(collection.Tag, " ") {
			if s == "" {
				continue
			}
			tags[s]++
		}
	}

	var phpTags = make([]subject.Tag, 0, len(tags))

	for name, count := range tags {
		name := name
		phpTags = append(phpTags, subject.Tag{
			Name:  &name,
			Count: count,
		})
	}

	sort.Slice(phpTags, func(i, j int) bool {
		if phpTags[i].Count != phpTags[j].Count {
			return phpTags[i].Count > phpTags[j].Count
		}

		return *phpTags[i].Name > *phpTags[j].Name
	})

	newTag, err := phpserialize.Marshal(lo.Slice(phpTags, 0, 30)) //nolint:gomnd
	if err != nil {
		return errgo.Wrap(err, "php.Marshal")
	}

	_, err = r.q.SubjectField.WithContext(ctx).Where(r.q.SubjectField.Sid.Eq(subjectID)).
		UpdateSimple(r.q.SubjectField.Tags.Value(newTag))

	return errgo.Wrap(err, "failed to update subject field")
}

func (r mysqlRepo) updateCollectionTime(obj *dao.SubjectCollection,
	t collection.SubjectCollection, at time.Time) error {
	switch t {
	case collection.SubjectCollectionAll:
		return errgo.Wrap(gerr.ErrInput, "can't set collection type to SubjectCollectionAll")
	case collection.SubjectCollectionWish:
		obj.WishTime = uint32(at.Unix())
	case collection.SubjectCollectionDone:
		obj.DoneTime = uint32(at.Unix())
	case collection.SubjectCollectionDoing:
		obj.DoingTime = uint32(at.Unix())
	case collection.SubjectCollectionDropped:
		obj.DroppedTime = uint32(at.Unix())
	case collection.SubjectCollectionOnHold:
		obj.OnHoldTime = uint32(at.Unix())
	default:
		return errgo.Wrap(gerr.ErrInput, fmt.Sprintln("invalid subject collection type", t))
	}
	return nil
}

func (r mysqlRepo) GetPersonCollection(
	ctx context.Context, userID model.UserID,
	cat collection.PersonCollectCategory, targetID model.PersonID,
) (collection.UserPersonCollection, error) {
	c, err := r.q.PersonCollect.WithContext(ctx).
		Where(r.q.PersonCollect.UserID.Eq(userID), r.q.PersonCollect.Category.Eq(string(cat)),
			r.q.PersonCollect.TargetID.Eq(targetID)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return collection.UserPersonCollection{}, gerr.ErrNotFound
		}
		return collection.UserPersonCollection{}, errgo.Wrap(err, "dal")
	}

	return collection.UserPersonCollection{
		ID:        c.ID,
		Category:  c.Category,
		TargetID:  c.TargetID,
		UserID:    c.UserID,
		CreatedAt: time.Unix(int64(c.CreatedTime), 0),
	}, nil
}

func (r mysqlRepo) AddPersonCollection(
	ctx context.Context, userID model.UserID,
	cat collection.PersonCollectCategory, targetID model.PersonID,
) error {
	collect := &dao.PersonCollect{
		UserID:      userID,
		Category:    string(cat),
		TargetID:    targetID,
		CreatedTime: uint32(time.Now().Unix()),
	}
	err := r.q.Transaction(func(tx *query.Query) error {
		switch cat {
		case collection.PersonCollectCategoryCharacter:
			if _, err := tx.Character.WithContext(ctx).Where(
				tx.Character.ID.Eq(targetID)).UpdateSimple(tx.Character.Collects.Add(1)); err != nil {
				r.log.Error("failed to update character collects", zap.Error(err))
				return err
			}
		case collection.PersonCollectCategoryPerson:
			if _, err := tx.Person.WithContext(ctx).Where(
				tx.Person.ID.Eq(targetID)).UpdateSimple(tx.Person.Collects.Add(1)); err != nil {
				r.log.Error("failed to update person collects", zap.Error(err))
				return err
			}
		}
		if err := tx.PersonCollect.WithContext(ctx).Create(collect); err != nil {
			r.log.Error("failed to create person collection record", zap.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		return errgo.Wrap(err, "dal")
	}
	return nil
}

func (r mysqlRepo) RemovePersonCollection(
	ctx context.Context, userID model.UserID,
	cat collection.PersonCollectCategory, targetID model.PersonID,
) error {
	err := r.q.Transaction(func(tx *query.Query) error {
		switch cat {
		case collection.PersonCollectCategoryCharacter:
			if _, err := tx.Character.WithContext(ctx).Where(
				tx.Character.ID.Eq(targetID)).UpdateSimple(tx.Character.Collects.Sub(1)); err != nil {
				r.log.Error("failed to update character collects", zap.Error(err))
				return err
			}
		case collection.PersonCollectCategoryPerson:
			if _, err := tx.Person.WithContext(ctx).Where(
				tx.Person.ID.Eq(targetID)).UpdateSimple(tx.Person.Collects.Sub(1)); err != nil {
				r.log.Error("failed to update person collects", zap.Error(err))
				return err
			}
		}
		_, err := tx.PersonCollect.WithContext(ctx).Where(
			tx.PersonCollect.UserID.Eq(userID),
			tx.PersonCollect.Category.Eq(string(cat)),
			tx.PersonCollect.TargetID.Eq(targetID),
		).Delete()
		if err != nil {
			r.log.Error("failed to delete person collection record", zap.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		return errgo.Wrap(err, "dal")
	}

	return nil
}

func (r mysqlRepo) CountPersonCollections(
	ctx context.Context,
	userID model.UserID,
	cat collection.PersonCollectCategory,
) (int64, error) {
	q := r.q.PersonCollect.WithContext(ctx).
		Where(r.q.PersonCollect.UserID.Eq(userID), r.q.PersonCollect.Category.Eq(string(cat)))

	c, err := q.Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}

	return c, nil
}

func (r mysqlRepo) ListPersonCollection(
	ctx context.Context,
	userID model.UserID,
	cat collection.PersonCollectCategory,
	limit, offset int,
) ([]collection.UserPersonCollection, error) {
	q := r.q.PersonCollect.WithContext(ctx).
		Order(r.q.PersonCollect.CreatedTime.Desc()).
		Where(r.q.PersonCollect.UserID.Eq(userID), r.q.PersonCollect.Category.Eq(string(cat))).Limit(limit).Offset(offset)

	collections, err := q.Find()
	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var results = make([]collection.UserPersonCollection, len(collections))
	for i, c := range collections {
		results[i] = collection.UserPersonCollection{
			ID:        c.ID,
			Category:  c.Category,
			TargetID:  c.TargetID,
			UserID:    c.UserID,
			CreatedAt: time.Unix(int64(c.CreatedTime), 0),
		}
	}

	return results, nil
}

func (r mysqlRepo) UpdateEpisodeCollection(
	ctx context.Context,
	userID model.UserID,
	subjectID model.SubjectID,
	episodeIDs []model.EpisodeID,
	collectionType collection.EpisodeCollection,
	at time.Time,
) (collection.UserSubjectEpisodesCollection, error) {
	table := r.q.EpCollection
	where := []gen.Condition{table.UserID.Eq(userID), table.SubjectID.Eq(subjectID)}

	d, err := table.WithContext(ctx).Where(where...).Take()
	if err != nil {
		// 章节表在用到时才会创建
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.createEpisodeCollection(ctx, userID, subjectID, episodeIDs, collectionType, at)
		}

		r.log.Error("failed to get episode collection record", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	e, err := deserializePhpEpStatus(d.Status)
	if err != nil {
		r.log.Error("failed to deserialize php-serialized bytes to go data", zap.Error(err))
		return nil, err
	}

	if updated := updateMysqlEpisodeCollection(e, episodeIDs, collectionType); !updated {
		return e.toModel(), nil
	}

	bytes, err := serializePhpEpStatus(e)
	if err != nil {
		return nil, err
	}

	_, err = table.WithContext(ctx).Where(where...).
		UpdateColumnSimple(table.Status.Value(bytes), table.UpdatedTime.Value(uint32(at.Unix())))
	if err != nil {
		return nil, errgo.Wrap(err, "EpCollection.UpdateColumnSimple")
	}

	return e.toModel(), nil
}

func (r mysqlRepo) createEpisodeCollection(
	ctx context.Context,
	userID model.UserID,
	subjectID model.SubjectID,
	episodeIDs []model.EpisodeID,
	collectionType collection.EpisodeCollection,
	at time.Time,
) (collection.UserSubjectEpisodesCollection, error) {
	var e = make(mysqlEpCollection, len(episodeIDs))
	updateMysqlEpisodeCollection(e, episodeIDs, collectionType)

	bytes, err := serializePhpEpStatus(e)
	if err != nil {
		return nil, err
	}

	table := r.q.EpCollection
	err = table.WithContext(ctx).Where(table.UserID.Eq(userID), table.SubjectID.Eq(subjectID)).Create(&dao.EpCollection{
		UserID:      userID,
		SubjectID:   subjectID,
		Status:      bytes,
		UpdatedTime: uint32(at.Unix()),
	})
	if err != nil {
		r.log.Error("failed to create episode collection record", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	return e.toModel(), nil
}

func updateMysqlEpisodeCollection(
	e mysqlEpCollection,
	episodeIDs []model.EpisodeID,
	collectionType collection.EpisodeCollection,
) bool {
	var updated bool

	if collectionType == collection.EpisodeCollectionNone {
		// remove episode collection
		for _, episodeID := range episodeIDs {
			_, ok := e[episodeID]
			if !ok {
				continue
			}

			delete(e, episodeID)
			updated = true
		}
	} else {
		for _, episodeID := range episodeIDs {
			v, ok := e[episodeID]
			if ok && v.Type == collectionType {
				continue
			}

			e[episodeID] = mysqlEpCollectionItem{EpisodeID: episodeID, Type: collectionType}
			updated = true
		}
	}

	return updated
}
