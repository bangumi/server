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
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/trim21/errgo"
	"github.com/trim21/go-phpserialize"
	"go.uber.org/zap"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

	err = r.updateUserTags(ctx, userID, subject, at, s)
	if err != nil {
		return errgo.Trace(err)
	}

	r.updateSubject(ctx, subject.ID)
	return nil
}

func (r mysqlRepo) updateUserTags(
	ctx context.Context,
	userID model.UserID,
	subject model.Subject,
	at time.Time,
	s *collection.Subject,
) error {
	return r.q.Transaction(func(q *query.Query) error {
		tx := q.WithContext(ctx)

		if (len(s.Tags())) == 0 {
			_, err := tx.TagList.Where(q.TagList.UID.Eq(userID),
				q.TagList.Mid.Eq(subject.ID),
				q.TagList.Cat.Eq(model.TagCatSubject)).Delete()
			if err != nil {
				return errgo.Trace(err)
			}
			return r.reCountSubjectTags(ctx, q, subject.ID)
		}

		tags, err := tx.TagIndex.Select().
			Where(q.TagIndex.Name.In(s.Tags()...), q.TagIndex.Cat.Eq(model.TagCatSubject)).Find()
		if err != nil {
			return errgo.Trace(err)
		}

		var existsTags = lo.SliceToMap(tags, func(item *dao.TagIndex) (string, bool) {
			return item.Name, true
		})

		var missingTags []string
		for _, tag := range s.Tags() {
			if !existsTags[tag] {
				missingTags = append(missingTags, tag)
			}
		}

		if len(missingTags) > 0 {
			err = tx.TagIndex.Create(lo.Map(missingTags, func(item string, index int) *dao.TagIndex {
				return &dao.TagIndex{
					Name:        item,
					Cat:         model.TagCatSubject,
					Type:        int8(subject.TypeID),
					Results:     1,
					CreatedTime: uint32(at.Unix()),
					UpdatedTime: uint32(at.Unix()),
				}
			})...)

			if err != nil {
				return errgo.Trace(err)
			}
		}

		tags, err = tx.TagIndex.Select().
			Where(q.TagIndex.Name.In(s.Tags()...), q.TagIndex.Cat.Eq(model.TagCatSubject)).Find()
		if err != nil {
			return errgo.Trace(err)
		}

		err = tx.TagList.Clauses(clause.OnConflict{DoNothing: true}).
			Create(lo.Map(tags, func(item *dao.TagIndex, index int) *dao.TagList {
				return &dao.TagList{
					Tid:         item.ID,
					UID:         s.User(),
					Cat:         model.TagCatSubject,
					Type:        subject.TypeID,
					Mid:         subject.ID,
					CreatedTime: uint32(at.Unix()),
				}
			})...)
		if err != nil {
			return errgo.Trace(err)
		}

		return r.reCountSubjectTags(ctx, q, subject.ID)
	})
}

func (r mysqlRepo) reCountSubjectTags(ctx context.Context, tx *query.Query, id model.SubjectID) error {
	tags, err := tx.WithContext(ctx).TagList.Select().
		Where(tx.TagList.Cat.Eq(model.TagCatSubject), tx.TagList.Mid.Eq(id)).Find()
	if err != nil {
		return err
	}

	db := tx.DB().WithContext(ctx)

	err = db.Exec(`
						update chii_tag_neue_index
							set tag_results = (
								select count(1)
                   from chii_tag_neue_list
                   where tlt_tid = chii_tag_neue_index.tag_id
							 )
						where tag_id in ?
	`, lo.Map(tags, func(item *dao.TagList, index int) uint32 {
		return item.Tid
	})).Error

	if err != nil {
		return errgo.Trace(err)
	}

	tagList, err := tx.WithContext(ctx).TagList.Preload(tx.TagList.Tag).
		Where(tx.TagList.Cat.Eq(model.TagCatSubject), tx.TagList.Mid.Eq(id)).Find()
	if err != nil {
		return errgo.Trace(err)
	}

	var count = make(map[string]int)

	for _, tag := range tagList {
		count[tag.Tag.Name]++
	}

	var phpTags = make([]subject.Tag, 0, len(count))

	for name, c := range count {
		phpTags = append(phpTags, subject.Tag{
			Name:  &name,
			Count: c,
		})
	}

	slices.SortFunc(phpTags, func(a, b subject.Tag) int {
		return cmp.Compare(a.Count, b.Count)
	})

	newTag, err := phpserialize.Marshal(lo.Slice(phpTags, 0, 30)) //nolint:gomnd
	if err != nil {
		return errgo.Wrap(err, "php.Marshal")
	}

	_, err = tx.WithContext(ctx).SubjectField.Where(r.q.SubjectField.Sid.Eq(id)).
		UpdateSimple(r.q.SubjectField.Tags.Value(newTag))

	return errgo.Wrap(err, "failed to update subject field")
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
	if err := r.reCountSubjectCollection(ctx, subjectID); err != nil {
		r.log.Error("failed to update collection counts", zap.Error(err))
	}
}

func (r mysqlRepo) reCountSubjectCollection(ctx context.Context, subjectID model.SubjectID) error {
	err := r.q.DB().WithContext(ctx).Exec(`
		update chii_subjects
		set subject_wish    = (select count(1)
													 from chii_subject_interests
													 where interest_subject_id = 1 and interest_type = subject_id),
				subject_collect = (select count(1)
													 from chii_subject_interests
													 where interest_subject_id = 2 and interest_type = subject_id),
				subject_doing   = (select count(1)
													 from chii_subject_interests
													 where interest_subject_id = 3 and interest_type = subject_id),
				subject_on_hold = (select count(1)
													 from chii_subject_interests
													 where interest_subject_id = 4 and interest_type = subject_id),
				subject_dropped = (select count(1)
													 from chii_subject_interests
													 where interest_subject_id = 5 and interest_type = subject_id)
		where chii_subjects.subject_id = ?
`, subjectID).Error
	return errgo.Trace(err)
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
