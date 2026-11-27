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
	"go.uber.org/zap"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/dal/utiltype"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/collections"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/dam"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/gmap"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/internal/pkg/serialize"
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

func lockSubjectCollection(
	ctx context.Context, tx *query.Query, userID model.UserID, subject model.Subject, allowCreate bool,
) (*dao.SubjectCollection, bool, error) {
	if allowCreate {
		obj := &dao.SubjectCollection{
			SubjectID:   subject.ID,
			SubjectType: subject.TypeID,
			UserID:      userID,
			Type:        uint8(collection.SubjectCollectionDoing),
		}
		// 键冲突时保留现有收藏。MySQL 驱动按实际变更计数，原值写回计为零行。
		result := tx.DB().WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{"interest_id": gorm.Expr("interest_id")}),
		}).Create(obj)
		if result.Error != nil {
			return nil, false, errgo.Wrap(result.Error, "ensure subject collection")
		}
		if result.RowsAffected == 1 {
			return obj, true, nil
		}
	}

	obj, err := tx.SubjectCollection.WithContext(ctx).
		Where(tx.SubjectCollection.UserID.Eq(userID), tx.SubjectCollection.SubjectID.Eq(subject.ID)).
		Clauses(clause.Locking{Strength: "UPDATE"}).Take()
	if err != nil {
		err = gerr.WrapGormError(err)
		if errors.Is(err, gerr.ErrNotFound) {
			return nil, false, gerr.ErrSubjectNotCollected
		}
		return nil, false, errgo.Wrap(err, "lock subject collection")
	}

	return obj, false, nil
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
	return r.updateOrCreateSubjectCollection(ctx, userID, subject, at, ip, update, false /* allowCreate */)
}

func (r mysqlRepo) UpdateOrCreateSubjectCollection(
	ctx context.Context,
	userID model.UserID,
	subject model.Subject,
	at time.Time,
	ip string,
	update func(ctx context.Context, s *collection.Subject) (*collection.Subject, error),
) error {
	return r.updateOrCreateSubjectCollection(ctx, userID, subject, at, ip, update, true /* allowCreate */)
}

//nolint:funlen
func (r mysqlRepo) updateOrCreateSubjectCollection(
	ctx context.Context,
	userID model.UserID,
	subject model.Subject,
	at time.Time,
	ip string,
	update func(ctx context.Context, s *collection.Subject) (*collection.Subject, error),
	allowCreate bool,
) error {
	var originalRate, updatedRate uint8
	err := r.q.Transaction(func(tx *query.Query) error {
		obj, created, err := lockSubjectCollection(ctx, tx, userID, subject, allowCreate)
		if err != nil {
			return errgo.Trace(err)
		}
		originalRate = obj.Rate
		collectionSubject, err := r.convertToSubjectCollection(obj)
		if err != nil {
			return errgo.Trace(err)
		}
		original := *collectionSubject
		s, err := update(ctx, collectionSubject)
		if err != nil {
			return errgo.Trace(err)
		}
		if updateErr := r.updateSubjectCollection(obj, &original, s, at, ip, created); updateErr != nil {
			return errgo.Trace(updateErr)
		}

		var affected []uint32
		if s.TagsSet() {
			var tags []string
			tags, affected, err = r.updateUserTags(ctx, tx, userID, subject, at, s.Tags())
			if err != nil {
				return errgo.Trace(err)
			}
			slices.Sort(tags)
			obj.Tag = strings.Join(tags, " ")
		}
		updatedRate = obj.Rate
		if _, writeErr := tx.SubjectCollection.WithContext(ctx).
			Where(tx.SubjectCollection.ID.Eq(obj.ID)).Select(field.ALL).
			Omit(tx.SubjectCollection.ID, tx.SubjectCollection.UserID, tx.SubjectCollection.SubjectID).
			Updates(obj); writeErr != nil {
			return errgo.Wrap(writeErr, "update subject collection")
		}
		if len(affected) != 0 {
			return r.reCountSubjectTags(ctx, tx, subject.ID, affected)
		}
		return nil
	})
	if err != nil {
		return errgo.Wrap(err, "save subject collection and tags")
	}

	r.updateSubject(ctx, subject.ID)

	if updatedRate != originalRate {
		if err := r.reCountSubjectRate(ctx, subject.ID, originalRate, updatedRate); err != nil {
			r.log.Error("failed to update collection counts", zap.Error(err), zap.Uint32("subject_id", subject.ID))
		}
	}

	return nil
}

func (r mysqlRepo) updateUserTags(ctx context.Context,
	q *query.Query,
	userID model.UserID, subject model.Subject,
	at time.Time, names []string) ([]string, []uint32, error) {
	// 先锁条目扩展行再读取标签，读快照才能包含同一条目上一次提交的修改。
	if _, err := q.SubjectField.WithContext(ctx).Select(q.SubjectField.Sid).
		Where(q.SubjectField.Sid.Eq(subject.ID)).Clauses(clause.Locking{Strength: "UPDATE"}).Find(); err != nil {
		return nil, nil, errgo.Wrap(err, "lock subject field")
	}

	tags, err := r.resolveSubjectTags(ctx, q, subject.TypeID, at, names)
	if err != nil {
		return nil, nil, errgo.Trace(err)
	}

	links := q.TagList.WithContext(ctx).Where(
		q.TagList.UID.Eq(userID), q.TagList.Mid.Eq(subject.ID),
		q.TagList.Cat.Eq(model.TagCatSubject),
	).Session(&gorm.Session{})
	oldLinks, err := links.Select(q.TagList.Tid).Find()
	if err != nil {
		return nil, nil, errgo.Wrap(err, "read collection tag links")
	}
	oldIDs := lo.Map(oldLinks, func(link *dao.TagList, _ int) uint32 { return link.Tid })
	newIDs := lo.Map(tags, func(tag *dao.TagIndex, _ int) uint32 { return tag.ID })
	slices.Sort(oldIDs)
	slices.Sort(newIDs)
	sanitized := slice.Map(tags, func(tag *dao.TagIndex) string { return tag.Name })
	if slices.Equal(oldIDs, newIDs) {
		return sanitized, nil, nil
	}

	if len(oldIDs) != 0 {
		if _, err := links.Delete(); err != nil {
			return nil, nil, errgo.Wrap(err, "delete collection tag links")
		}
	}
	if len(newIDs) != 0 {
		rows := lo.Map(newIDs, func(id uint32, _ int) *dao.TagList {
			return &dao.TagList{
				Tid:         id,
				UID:         userID,
				Cat:         model.TagCatSubject,
				Type:        subject.TypeID,
				Mid:         subject.ID,
				CreatedTime: uint32(at.Unix()),
			}
		})
		if err := q.TagList.WithContext(ctx).Create(rows...); err != nil {
			return nil, nil, errgo.Wrap(err, "insert collection tag links")
		}
	}
	return sanitized, lo.Uniq(append(oldIDs, newIDs...)), nil
}

func (r mysqlRepo) resolveSubjectTags(ctx context.Context, q *query.Query,
	subjectType model.SubjectType, at time.Time, names []string) ([]*dao.TagIndex, error) {
	if len(names) == 0 {
		return nil, nil
	}
	names = lo.UniqBy(names, strings.ToLower)
	tagIndex := q.TagIndex.WithContext(ctx)
	find := func() (map[string]*dao.TagIndex, error) {
		tags, err := tagIndex.
			Where(q.TagIndex.Name.In(names...), q.TagIndex.Cat.Eq(model.TagCatSubject),
				q.TagIndex.Type.Eq(subjectType)).Order(q.TagIndex.ID.Asc()).Find()
		if err != nil {
			return nil, err
		}
		byName := make(map[string]*dao.TagIndex, len(tags))
		for _, tag := range tags {
			key := strings.ToLower(tag.Name)
			if byName[key] == nil {
				byName[key] = tag
			}
		}
		return byName, nil
	}

	byName, err := find()
	if err != nil {
		return nil, errgo.Wrap(err, "find subject tags")
	}

	missingTags := lo.Filter(names, func(name string, _ int) bool { return byName[strings.ToLower(name)] == nil })

	if len(missingTags) > 0 {
		r.log.Info("create missing tags", zap.Strings("missing_tags", missingTags))
		err = tagIndex.Create(lo.Map(missingTags, func(item string, _ int) *dao.TagIndex {
			return &dao.TagIndex{
				Name:        item,
				Cat:         model.TagCatSubject,
				Type:        subjectType,
				Results:     1,
				CreatedTime: uint32(at.Unix()),
				UpdatedTime: uint32(at.Unix()),
			}
		})...)

		if err != nil {
			return nil, errgo.Wrap(err, "create subject tags")
		}
		byName, err = find()
		if err != nil {
			return nil, errgo.Wrap(err, "find subject tags after creation")
		}
	}
	// SQL 可能把 cafe 和 café 都查出来，这里只取请求中匹配的标签。
	tags := slice.MapFilter(names, func(name string) (*dao.TagIndex, bool) {
		tag := byName[strings.ToLower(name)]
		return tag, tag != nil
	})
	if len(tags) != len(names) {
		return nil, fmt.Errorf("resolve subject tags: matched %d of %d names", len(tags), len(names))
	}
	return tags, nil
}

func (r mysqlRepo) reCountSubjectTags(ctx context.Context, tx *query.Query,
	subjectID model.SubjectID, affected []uint32) error {
	db := tx.DB().WithContext(ctx)

	err := db.Exec(`
			update chii_tag_neue_index as ti
				set tag_results = (
					select count(1)
						 from chii_tag_neue_list as tl
						 where tl.tlt_cat = ti.tag_cat AND tl.tlt_tid = ti.tag_id and tl.tlt_type = ti.tag_type
				 )
			where ti.tag_cat = ? AND ti.tag_id IN ?
	`, model.TagCatSubject, affected).Error
	if err != nil {
		return errgo.Trace(err)
	}

	tagList, err := tx.WithContext(ctx).TagList.Preload(tx.TagList.Tag).
		Where(tx.TagList.Cat.Eq(model.TagCatSubject), tx.TagList.Mid.Eq(subjectID)).Find()
	if err != nil {
		return errgo.Trace(err)
	}

	var count = make(map[string]uint)
	var countMap = make(map[string]uint)

	for _, tag := range tagList {
		if !dam.ValidateTag(tag.Tag.Name) {
			continue
		}

		count[tag.Tag.Name]++
		countMap[tag.Tag.Name] = uint(tag.Tag.Results)
	}

	var phpTags = make([]subject.Tag, 0, len(count))

	for name, c := range count {
		phpTags = append(phpTags, subject.Tag{
			Name:       lo.ToPtr(name),
			Count:      c,
			TotalCount: countMap[name],
		})
	}

	slices.SortFunc(phpTags, func(a, b subject.Tag) int {
		return -cmp.Compare(a.Count, b.Count)
	})

	newTag, err := serialize.Encode(lo.Slice(phpTags, 0, 30)) //nolint:mnd
	if err != nil {
		return errgo.Wrap(err, "php.Marshal")
	}

	_, err = tx.WithContext(ctx).SubjectField.Where(r.q.SubjectField.Sid.Eq(subjectID)).
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
	obj.EpStatus = s.Eps()
	obj.VolStatus = s.Vols()
	obj.Rate = s.Rate()
	obj.Private = uint8(s.Privacy())
	obj.Type = uint8(s.TypeID())

	if isNew || s.TypeID() != original.TypeID() {
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
		Where(r.q.SubjectCollection.UserID.Eq(userID),
			r.q.SubjectCollection.Type.Neq(0),
		)

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
		Where(
			r.q.SubjectCollection.UserID.Eq(userID),
			r.q.SubjectCollection.Type.Neq(0),
		).Limit(limit).Offset(offset)

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
		Where(
			r.q.SubjectCollection.UserID.Eq(userID),
			r.q.SubjectCollection.SubjectID.Eq(subjectID),
			r.q.SubjectCollection.Type.Neq(0),
		).Take()
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

	e, err := deserializeEpStatus(d.Status)
	if err != nil {
		return nil, err
	}

	return e.toModel(), nil
}

func (r mysqlRepo) updateSubject(ctx context.Context, subjectID model.SubjectID) {
	if err := r.reCountSubjectCollection(ctx, subjectID); err != nil {
		r.log.Error("failed to update collection counts", zap.Error(err), zap.Uint32("subject_id", subjectID))
	}
}

func (r mysqlRepo) reCountSubjectCollection(ctx context.Context, subjectID model.SubjectID) error {
	var counts []struct {
		Type  uint8  `gorm:"type"`
		Total uint32 `gorm:"total"`
	}

	return r.q.Transaction(func(tx *query.Query) error {
		err := tx.DB().WithContext(ctx).Raw(`
			select interest_type as type, count(interest_type) as total
			  from chii_subject_interests
				where interest_subject_id = ? and interest_type != 0
			group by interest_type
		`, subjectID).Scan(&counts).Error
		if err != nil {
			return errgo.Wrap(err, "dal")
		}

		var updater = make([]field.AssignExpr, 0, 5)
		for _, count := range counts {
			switch collection.SubjectCollection(count.Type) {
			case collection.SubjectCollectionAll:
				continue

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

		_, err = tx.Subject.WithContext(ctx).Where(r.q.Subject.ID.Eq(subjectID)).UpdateSimple(updater...)
		if err != nil {
			return errgo.Wrap(err, "dal")
		}

		return nil
	})
}

//nolint:mnd,gocyclo
func (r mysqlRepo) reCountSubjectRate(ctx context.Context, subjectID model.SubjectID, before uint8, after uint8) error {
	return r.q.Transaction(func(tx *query.Query) error {
		var counts = make(map[uint8]uint32, 2)

		for _, rate := range []uint8{before, after} {
			var count uint32
			if rate != 0 {
				err := tx.DB().WithContext(ctx).Raw(`
			select count(*) from chii_subject_interests
			where interest_subject_id = ? and interest_private = 0 and interest_rate = ?
		`, subjectID, rate).Scan(&count).Error
				if err != nil {
					return errgo.Wrap(err, "dal")
				}

				counts[rate] = count
			}
		}

		var updater = make([]field.AssignExpr, 0, 2)
		for rate, total := range counts {
			switch rate {
			case 0:
				continue
			case 1:
				updater = append(updater, tx.SubjectField.Rate1.Value(total))
			case 2:
				updater = append(updater, tx.SubjectField.Rate2.Value(total))
			case 3:
				updater = append(updater, tx.SubjectField.Rate3.Value(total))
			case 4:
				updater = append(updater, tx.SubjectField.Rate4.Value(total))
			case 5:
				updater = append(updater, tx.SubjectField.Rate5.Value(total))
			case 6:
				updater = append(updater, tx.SubjectField.Rate6.Value(total))
			case 7:
				updater = append(updater, tx.SubjectField.Rate7.Value(total))
			case 8:
				updater = append(updater, tx.SubjectField.Rate8.Value(total))
			case 9:
				updater = append(updater, tx.SubjectField.Rate9.Value(total))
			case 10:
				updater = append(updater, tx.SubjectField.Rate10.Value(total))
			}
		}

		_, err := tx.SubjectField.WithContext(ctx).Where(r.q.SubjectField.Sid.Eq(subjectID)).UpdateSimple(updater...)
		if err != nil {
			return errgo.Wrap(err, "dal")
		}

		return nil
	})
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

	e, err := deserializeEpStatus(d.Status)
	if err != nil {
		r.log.Error("failed to deserialize php-serialized bytes to go data", zap.Error(err))
		return nil, err
	}

	if updated := updateMysqlEpisodeCollection(e, episodeIDs, collectionType, at); !updated {
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
	updateMysqlEpisodeCollection(e, episodeIDs, collectionType, at)

	bytes, err := serializePhpEpStatus(e)
	if err != nil {
		return nil, err
	}

	table := r.q.EpCollection
	err = table.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Where(table.UserID.Eq(userID), table.SubjectID.Eq(subjectID)).Create(&dao.EpCollection{
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
	now time.Time,
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
		return updated
	}

	for _, episodeID := range episodeIDs {
		v, ok := e[episodeID]
		if ok {
			if v.Type == collectionType {
				continue
			}

			v.Type = collectionType
			v.UpdatedAt = gmap.SafeAssign(v.UpdatedAt, collectionType, now.Unix())
			updated = true
			continue
		}

		e[episodeID] = mysqlEpCollectionItem{
			EpisodeID: episodeID,
			Type:      collectionType,
			UpdatedAt: map[collection.EpisodeCollection]int64{collectionType: now.Unix()},
		}
		updated = true
	}

	return updated
}
