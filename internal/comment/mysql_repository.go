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

package comment

import (
	"context"
	"errors"
	"reflect"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.CommentRepo, error) {
	return mysqlRepo{q: q, log: log.Named("comment.mysqlRepo")}, nil
}

func (r mysqlRepo) Get(
	ctx context.Context, commentType domain.CommentType, id model.CommentID,
) (model.Comment, error) {
	var (
		comment interface{}
		err     error
	)
	switch commentType {
	case domain.CommentTypeGroupTopic:
		comment, err = r.q.GroupTopicComment.WithContext(ctx).Where(r.q.GroupTopicComment.ID.Eq(id)).First()
	case domain.CommentTypeSubjectTopic:
		comment, err = r.q.SubjectTopicComment.WithContext(ctx).Where(r.q.SubjectTopicComment.ID.Eq(id)).First()
	case domain.CommentIndex:
		comment, err = r.q.IndexComment.WithContext(ctx).Where(r.q.IndexComment.ID.Eq(id)).First()
	case domain.CommentCharacter:
		comment, err = r.q.CharacterComment.WithContext(ctx).Where(r.q.CharacterComment.ID.Eq(id)).First()
	case domain.CommentPerson:
		comment, err = r.q.CharacterComment.WithContext(ctx).Where(r.q.CharacterComment.ID.Eq(id)).First()
	case domain.CommentEpisode:
		comment, err = r.q.EpisodeComment.WithContext(ctx).Where(r.q.EpisodeComment.ID.Eq(id)).First()
	default:
		return model.Comment{}, errUnsupportCommentType
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Comment{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return model.Comment{}, errgo.Wrap(err, "dal")
	}

	return convertDao(comment)
}

//nolint:gocyclo
func (r mysqlRepo) GetByIDs(
	ctx context.Context, commentType domain.CommentType, ids ...model.CommentID,
) (map[model.CommentID]model.Comment, error) {
	var (
		rawComments interface{}
		commentMap  = make(map[model.CommentID]model.Comment)
		err         error
	)
	switch commentType {
	case domain.CommentTypeGroupTopic:
		rawComments, err = r.q.GroupTopicComment.WithContext(ctx).Where(r.q.GroupTopicComment.ID.In(ids...)).Find()
	case domain.CommentTypeSubjectTopic:
		rawComments, err = r.q.SubjectTopicComment.WithContext(ctx).Where(r.q.SubjectTopicComment.ID.In(ids...)).Find()
	case domain.CommentIndex:
		rawComments, err = r.q.IndexComment.WithContext(ctx).Where(r.q.IndexComment.ID.In(ids...)).Find()
	case domain.CommentCharacter:
		rawComments, err = r.q.CharacterComment.WithContext(ctx).Where(r.q.CharacterComment.ID.In(ids...)).Find()
	case domain.CommentPerson:
		rawComments, err = r.q.CharacterComment.WithContext(ctx).Where(r.q.CharacterComment.ID.In(ids...)).Find()
	case domain.CommentEpisode:
		rawComments, err = r.q.EpisodeComment.WithContext(ctx).Where(r.q.EpisodeComment.ID.In(ids...)).Find()
	default:
		return nil, errUnsupportCommentType
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	comments, err := convertModelComments(rawComments)
	if err != nil {
		return nil, err
	}
	for _, v := range comments {
		commentMap[v.ID] = v
	}
	return commentMap, nil
}

//nolint:gocyclo
func (r mysqlRepo) GetByRelateIDs(
	ctx context.Context, commentType domain.CommentType, ids ...model.CommentID,
) (map[model.CommentID][]model.Comment, error) {
	var (
		rawComments interface{}
		commentMap  = make(map[model.CommentID][]model.Comment)
		err         error
	)
	switch commentType {
	case domain.CommentTypeGroupTopic:
		rawComments, err = r.q.GroupTopicComment.WithContext(ctx).Where(r.q.GroupTopicComment.Related.In(ids...)).Find()
	case domain.CommentTypeSubjectTopic:
		rawComments, err = r.q.SubjectTopicComment.WithContext(ctx).Where(r.q.SubjectTopicComment.Related.In(ids...)).Find()
	case domain.CommentIndex:
		rawComments, err = r.q.IndexComment.WithContext(ctx).Where(r.q.IndexComment.Related.In(ids...)).Find()
	case domain.CommentCharacter:
		rawComments, err = r.q.CharacterComment.WithContext(ctx).Where(r.q.CharacterComment.Related.In(ids...)).Find()
	case domain.CommentPerson:
		rawComments, err = r.q.CharacterComment.WithContext(ctx).Where(r.q.CharacterComment.Related.In(ids...)).Find()
	case domain.CommentEpisode:
		rawComments, err = r.q.EpisodeComment.WithContext(ctx).Where(r.q.EpisodeComment.Related.In(ids...)).Find()
	default:
		return nil, errUnsupportCommentType
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	comments, err := convertModelComments(rawComments)
	if err != nil {
		return nil, err
	}
	for _, v := range comments {
		if m, ok := commentMap[v.ID]; !ok {
			commentMap[v.Related] = append(m, v)
		} else {
			commentMap[v.Related] = []model.Comment{v}
		}
	}
	return commentMap, nil
}

var errUnsupportCommentType = errors.New("comment type not support")

func convertDao(in interface{}) (model.Comment, error) {
	switch v := in.(type) {
	case *dao.SubjectTopicComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         model.UserID(v.UID),
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.CreatedTime), 0),
			Content:     v.Content,
		}, nil
	case *dao.GroupTopicComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         model.UserID(v.UID),
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.CreatedTime), 0),
			Content:     v.Content,
		}, nil
	case *dao.IndexComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         model.UserID(v.UID),
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.CreatedTime), 0),
			Content:     v.Content,
		}, nil
	case *dao.EpisodeComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         model.UserID(v.UID),
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.CreatedTime), 0),
			Content:     v.Content,
		}, nil
	case *dao.CharacterComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         model.UserID(v.UID),
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.CreatedTime), 0),
			Content:     v.Content,
		}, nil
	case *dao.PersonComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         model.UserID(v.UID),
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.CreatedTime), 0),
			Content:     v.Content,
		}, nil
	default:
		return model.Comment{}, errUnsupportCommentType
	}
}

func (r mysqlRepo) Count(ctx context.Context, commentType domain.CommentType, id uint32) (int64, error) {
	var (
		count int64
		err   error
	)
	switch commentType {
	case domain.CommentTypeGroupTopic:
		count, err = r.q.GroupTopicComment.WithContext(ctx).
			Where(r.q.GroupTopicComment.Related.Eq(0)).
			Where(r.q.GroupTopicComment.MentionedID.Eq(id)).Count()
	case domain.CommentTypeSubjectTopic:
		count, err = r.q.SubjectTopicComment.WithContext(ctx).
			Where(r.q.SubjectTopicComment.Related.Eq(0)).
			Where(r.q.SubjectTopicComment.MentionedID.Eq(id)).Count()
	case domain.CommentIndex:
		count, err = r.q.IndexComment.WithContext(ctx).
			Where(r.q.IndexComment.Related.Eq(0)).
			Where(r.q.IndexComment.MentionedID.Eq(id)).Count()
	case domain.CommentCharacter:
		count, err = r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.Eq(0)).
			Where(r.q.CharacterComment.MentionedID.Eq(id)).Count()
	case domain.CommentPerson:
		count, err = r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.Eq(0)).
			Where(r.q.CharacterComment.MentionedID.Eq(id)).Count()
	case domain.CommentEpisode:
		count, err = r.q.EpisodeComment.WithContext(ctx).
			Where(r.q.EpisodeComment.Related.Eq(0)).
			Where(r.q.EpisodeComment.MentionedID.Eq(id)).Count()
	default:
		return 0, errUnsupportCommentType
	}
	if err != nil {
		return count, errgo.Wrap(err, "dal")
	}
	return count, nil
}

func (r mysqlRepo) List(
	ctx context.Context, commentType domain.CommentType, id uint32, limit int, offset int,
) ([]model.Comment, error) {
	var (
		comments interface{}
		err      error
	)
	switch commentType {
	case domain.CommentTypeGroupTopic:
		comments, err = r.q.GroupTopicComment.WithContext(ctx).
			Where(r.q.GroupTopicComment.Related.Eq(0)).
			Where(r.q.GroupTopicComment.MentionedID.Eq(id)).Offset(offset).Limit(limit).Find()
	case domain.CommentTypeSubjectTopic:
		comments, err = r.q.SubjectTopicComment.WithContext(ctx).
			Where(r.q.SubjectTopicComment.Related.Eq(0)).
			Where(r.q.SubjectTopicComment.MentionedID.Eq(id)).Offset(offset).Limit(limit).Find()
	case domain.CommentIndex:
		comments, err = r.q.IndexComment.WithContext(ctx).
			Where(r.q.IndexComment.Related.Eq(0)).
			Where(r.q.IndexComment.MentionedID.Eq(id)).Offset(offset).Limit(limit).Find()
	case domain.CommentCharacter:
		comments, err = r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.Eq(0)).
			Where(r.q.CharacterComment.MentionedID.Eq(id)).Offset(offset).Limit(limit).Find()
	case domain.CommentPerson:
		comments, err = r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.Eq(0)).
			Where(r.q.CharacterComment.MentionedID.Eq(id)).Offset(offset).Limit(limit).Find()
	case domain.CommentEpisode:
		comments, err = r.q.EpisodeComment.WithContext(ctx).
			Where(r.q.EpisodeComment.Related.Eq(0)).
			Where(r.q.EpisodeComment.MentionedID.Eq(id)).Offset(offset).Limit(limit).Find()
	default:
		return nil, errUnsupportCommentType
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}
	result, err := convertModelComments(comments)
	if err != nil {
		return nil, errgo.Wrap(err, "convert user")
	}
	return result, nil
}

var errInputNilComments = errors.New("input nil comments")
var errInputInvalidComments = errors.New("input invalid comments")

func convertModelComments(in interface{}) ([]model.Comment, error) {
	if !reflect.ValueOf(in).IsValid() {
		return nil, errInputNilComments
	}
	if reflect.TypeOf(in).Kind() == reflect.Slice {
		s := reflect.ValueOf(in)
		comments := make([]model.Comment, s.Len())
		for i := 0; i < s.Len(); i++ {
			if comment, err := convertDao(s.Index(i).Interface()); err == nil {
				comments[i] = comment
			} else {
				return comments, err
			}
		}
		return comments, nil
	}
	return nil, errInputInvalidComments
}
