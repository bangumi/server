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
	"sort"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.CommentRepo, error) {
	return mysqlRepo{q: q, log: log.Named("comment.mysqlRepo")}, nil
}

func (r mysqlRepo) Get(ctx context.Context, commentType domain.CommentType, id model.CommentID) (model.Comment, error) {
	var comment model.Commenter
	var err error
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

	return model.Comment{
		CreatedAt:   comment.CreateAt(),
		Content:     comment.GetContent(),
		CreatorID:   comment.CreatorID(),
		State:       comment.GetState(),
		ID:          comment.CommentID(),
		SubComments: nil,
	}, nil
}

func (r mysqlRepo) GetByRelateIDs(
	ctx context.Context, commentType domain.CommentType, ids ...model.CommentID,
) (map[model.CommentID][]model.Comment, error) {
	return map[model.CommentID][]model.Comment{}, nil
	// var (
	// 	rawComments []model.Commenter
	// 	err         error
	// )
	// switch commentType {
	// case domain.CommentTypeGroupTopic:
	// 	rawComments, err = wrapCommentDao(r.q.GroupTopicComment.WithContext(ctx).
	// 		Where(r.q.GroupTopicComment.Related.In(ids...)).Find())
	// case domain.CommentTypeSubjectTopic:
	// 	rawComments, err = wrapCommentDao(r.q.SubjectTopicComment.WithContext(ctx).
	// 		Where(r.q.SubjectTopicComment.Related.In(ids...)).Find())
	// case domain.CommentIndex:
	// 	rawComments, err = wrapCommentDao(r.q.IndexComment.WithContext(ctx).
	// 		Where(r.q.IndexComment.Related.In(ids...)).Find())
	// case domain.CommentCharacter:
	// 	rawComments, err = wrapCommentDao(r.q.CharacterComment.WithContext(ctx).
	// 		Where(r.q.CharacterComment.Related.In(ids...)).Find())
	// case domain.CommentPerson:
	// 	rawComments, err = wrapCommentDao(r.q.CharacterComment.WithContext(ctx).
	// 		Where(r.q.CharacterComment.Related.In(ids...)).Find())
	// case domain.CommentEpisode:
	// 	rawComments, err = wrapCommentDao(r.q.EpisodeComment.WithContext(ctx).
	// 		Where(r.q.EpisodeComment.Related.In(ids...)).Find())
	// default:
	// 	return nil, errUnsupportCommentType
	// }
	// if err != nil {
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		return nil, domain.ErrNotFound
	// 	}
	//
	// 	r.log.Error("unexpected error happened", zap.Error(err))
	// 	return nil, errgo.Wrap(err, "dal")
	// }
	//
	// return convertModelComments(rawComments), nil
}

var errUnsupportCommentType = errors.New("comment type not support")

func (r mysqlRepo) Count(ctx context.Context, commentType domain.CommentType, id uint32) (int64, error) {
	var count int64
	var err error
	switch commentType {
	case domain.CommentTypeGroupTopic:
		count, err = r.q.GroupTopicComment.WithContext(ctx).
			Where(r.q.GroupTopicComment.Related.Eq(0), r.q.GroupTopicComment.MentionedID.Eq(id)).Count()
	case domain.CommentTypeSubjectTopic:
		count, err = r.q.SubjectTopicComment.WithContext(ctx).
			Where(r.q.SubjectTopicComment.Related.Eq(0), r.q.SubjectTopicComment.MentionedID.Eq(id)).Count()
	case domain.CommentIndex:
		count, err = r.q.IndexComment.WithContext(ctx).
			Where(r.q.IndexComment.Related.Eq(0), r.q.IndexComment.MentionedID.Eq(id)).Count()
	case domain.CommentCharacter:
		count, err = r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.Eq(0), r.q.CharacterComment.MentionedID.Eq(id)).Count()
	case domain.CommentPerson:
		count, err = r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.Eq(0), r.q.CharacterComment.MentionedID.Eq(id)).Count()
	case domain.CommentEpisode:
		count, err = r.q.EpisodeComment.WithContext(ctx).
			Where(r.q.EpisodeComment.Related.Eq(0), r.q.EpisodeComment.MentionedID.Eq(id)).Count()
	default:
		return 0, errUnsupportCommentType
	}
	if err != nil {
		return count, errgo.Wrap(err, "dal")
	}
	return count, nil
}

func (r mysqlRepo) List(
	ctx context.Context, commentType domain.CommentType, topicID uint32, limit int, offset int,
) ([]model.Comment, error) {
	var comments []model.Commenter
	var err error

	switch commentType {
	case domain.CommentTypeGroupTopic:
		comments, err = wrapCommentDao(r.q.GroupTopicComment.WithContext(ctx).
			Where(r.q.GroupTopicComment.Related.Eq(0), r.q.GroupTopicComment.MentionedID.Eq(topicID)).
			Offset(offset).Limit(limit).Order(r.q.GroupTopicComment.CreatedTime, r.q.GroupTopicComment.ID).Find())
	case domain.CommentTypeSubjectTopic:
		comments, err = wrapCommentDao(r.q.SubjectTopicComment.WithContext(ctx).
			Where(r.q.SubjectTopicComment.Related.Eq(0), r.q.SubjectTopicComment.MentionedID.Eq(topicID)).
			Offset(offset).Limit(limit).Order(r.q.SubjectTopicComment.CreatedTime, r.q.SubjectTopicComment.ID).Find())
	case domain.CommentIndex:
		comments, err = wrapCommentDao(r.q.IndexComment.WithContext(ctx).
			Where(r.q.IndexComment.Related.Eq(0), r.q.IndexComment.MentionedID.Eq(topicID)).
			Offset(offset).Limit(limit).Order(r.q.IndexComment.CreatedTime, r.q.IndexComment.ID).Find())
	case domain.CommentCharacter:
		comments, err = wrapCommentDao(r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.Eq(0), r.q.CharacterComment.MentionedID.Eq(topicID)).
			Offset(offset).Limit(limit).Order(r.q.CharacterComment.CreatedTime, r.q.CharacterComment.ID).Find())
	case domain.CommentPerson:
		comments, err = wrapCommentDao(r.q.PersonComment.WithContext(ctx).
			Where(r.q.PersonComment.Related.Eq(0), r.q.PersonComment.MentionedID.Eq(topicID)).
			Offset(offset).Limit(limit).Order(r.q.PersonComment.CreatedTime, r.q.PersonComment.ID).Find())
	case domain.CommentEpisode:
		comments, err = wrapCommentDao(r.q.EpisodeComment.WithContext(ctx).
			Where(r.q.EpisodeComment.Related.Eq(0), r.q.EpisodeComment.MentionedID.Eq(topicID)).
			Offset(offset).Limit(limit).Order(r.q.EpisodeComment.CreatedTime, r.q.EpisodeComment.ID).Find())
	default:
		return nil, errUnsupportCommentType
	}

	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	return convertModelComments(comments), nil
}

func convertModelComments(in []model.Commenter) []model.Comment {
	commentMap := map[model.CommentID]model.Comment{}
	for _, comment := range in {
		if comment.IsSubComment() {
			continue
		}

		commentMap[comment.CommentID()] = model.Comment{
			CreatedAt:   comment.CreateAt(),
			Content:     comment.GetContent(),
			CreatorID:   comment.CreatorID(),
			State:       comment.GetState(),
			ID:          comment.CommentID(),
			SubComments: make([]model.SubComment, 0, 4),
		}
	}

	for _, comment := range in {
		if !comment.IsSubComment() {
			continue
		}

		parent := commentMap[comment.RelatedTo()]

		parent.SubComments = append(parent.SubComments, model.SubComment{
			CreatedAt:   comment.CreateAt(),
			Content:     comment.GetContent(),
			CreatorID:   comment.CreatorID(),
			Related:     comment.RelatedTo(),
			State:       comment.GetState(),
			ID:          comment.CommentID(),
			MentionedID: comment.GetMentionedID(),
		})

		commentMap[comment.RelatedTo()] = parent
	}

	comments := generic.MapValues(commentMap)

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	return comments
}
