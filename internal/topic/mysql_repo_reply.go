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

package topic

import (
	"context"
	"errors"
	"sort"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/gmap"
	"github.com/bangumi/server/internal/pkg/generic/slice"
)

var errUnsupportedCommentType = errors.New("comment type not support")

func (r mysqlRepo) CountReplies(ctx context.Context, commentType domain.CommentType, id model.TopicID) (int64, error) {
	// 帖子中第一个回复是楼主的发帖，后续的才是回复。

	var count int64
	var err error
	switch commentType {
	case domain.CommentTypeGroupTopic:
		count, err = r.q.GroupTopicComment.WithContext(ctx).
			Where(r.q.GroupTopicComment.Related.Eq(0), r.q.GroupTopicComment.TopicID.Eq(uint32(id))).Count()
		count--
	case domain.CommentTypeSubjectTopic:
		count, err = r.q.SubjectTopicComment.WithContext(ctx).
			Where(r.q.SubjectTopicComment.Related.Eq(0), r.q.SubjectTopicComment.TopicID.Eq(uint32(id))).Count()
		count--
	case domain.CommentIndex:
		count, err = r.q.IndexComment.WithContext(ctx).
			Where(r.q.IndexComment.Related.Eq(0), r.q.IndexComment.TopicID.Eq(uint32(id))).Count()
	case domain.CommentCharacter:
		count, err = r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.Eq(0), r.q.CharacterComment.TopicID.Eq(uint32(id))).Count()
	case domain.CommentPerson:
		count, err = r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.Eq(0), r.q.CharacterComment.TopicID.Eq(uint32(id))).Count()
	case domain.CommentEpisode:
		count, err = r.q.EpisodeComment.WithContext(ctx).
			Where(r.q.EpisodeComment.Related.Eq(0), r.q.EpisodeComment.TopicID.Eq(uint32(id))).Count()
	default:
		return 0, errUnsupportedCommentType
	}
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}
	return count, nil
}

func (r mysqlRepo) ListReplies(
	ctx context.Context, commentType domain.CommentType, id model.TopicID, limit int, offset int,
) ([]model.Comment, error) {
	commentMap, err := r.getParentComments(ctx, commentType, id, limit, offset)
	if err != nil {
		return nil, err
	}

	comments, err := r.getSubComments(ctx, commentType, id, gmap.Keys(commentMap)...)
	if err != nil {
		r.log.Error("failed to get sub replies")
		return nil, err
	}

	for _, comment := range comments {
		parent := commentMap[comment.RelatedTo()]

		parent.SubComments = append(parent.SubComments, model.SubComment{
			ID:        comment.CommentID(),
			CreatorID: comment.CreatorID(),
			CreatedAt: comment.CreateAt(),
			Content:   comment.GetContent(),
			Related:   comment.RelatedTo(),
			State:     model.CommentState(comment.GetState()),
			// TopicID:  comment.GetTopicID(),
		})

		commentMap[parent.ID] = parent
	}

	data := gmap.Values(commentMap)
	sort.Slice(data, func(i, j int) bool {
		if data[i].CreatedAt.Equal(data[j].CreatedAt) {
			return data[i].ID < data[j].ID
		}

		return data[i].CreatedAt.Before(data[j].CreatedAt)
	})

	return data, nil
}

func (r mysqlRepo) getParentComments(
	ctx context.Context, commentType domain.CommentType, id model.TopicID, limit int, offset int,
) (map[model.CommentID]model.Comment, error) {
	var comments []mysqlComment
	var err error

	switch commentType {
	case domain.CommentTypeGroupTopic:
		offset++
		comments, err = wrapCommentDao(r.q.GroupTopicComment.WithContext(ctx).
			Where(r.q.GroupTopicComment.Related.Eq(0), r.q.GroupTopicComment.TopicID.Eq(uint32(id))).
			Offset(offset).Limit(limit).Order(r.q.GroupTopicComment.CreatedTime, r.q.GroupTopicComment.ID).Find())
	case domain.CommentTypeSubjectTopic:
		offset++
		comments, err = wrapCommentDao(r.q.SubjectTopicComment.WithContext(ctx).
			Where(r.q.SubjectTopicComment.Related.Eq(0), r.q.SubjectTopicComment.TopicID.Eq(uint32(id))).
			Offset(offset).Limit(limit).Order(r.q.SubjectTopicComment.CreatedTime, r.q.SubjectTopicComment.ID).Find())
	case domain.CommentIndex:
		comments, err = wrapCommentDao(r.q.IndexComment.WithContext(ctx).
			Where(r.q.IndexComment.Related.Eq(0), r.q.IndexComment.TopicID.Eq(uint32(id))).
			Offset(offset).Limit(limit).Order(r.q.IndexComment.CreatedTime, r.q.IndexComment.ID).Find())
	case domain.CommentCharacter:
		comments, err = wrapCommentDao(r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.Eq(0), r.q.CharacterComment.TopicID.Eq(uint32(id))).
			Offset(offset).Limit(limit).Order(r.q.CharacterComment.CreatedTime, r.q.CharacterComment.ID).Find())
	case domain.CommentPerson:
		comments, err = wrapCommentDao(r.q.PersonComment.WithContext(ctx).
			Where(r.q.PersonComment.Related.Eq(0), r.q.PersonComment.TopicID.Eq(uint32(id))).
			Offset(offset).Limit(limit).Order(r.q.PersonComment.CreatedTime, r.q.PersonComment.ID).Find())
	case domain.CommentEpisode:
		comments, err = wrapCommentDao(r.q.EpisodeComment.WithContext(ctx).
			Where(r.q.EpisodeComment.Related.Eq(0), r.q.EpisodeComment.TopicID.Eq(uint32(id))).
			Offset(offset).Limit(limit).Order(r.q.EpisodeComment.CreatedTime, r.q.EpisodeComment.ID).Find())
	default:
		return nil, errUnsupportedCommentType
	}

	if err != nil {
		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	parents := slice.Map(comments, func(comment mysqlComment) model.Comment {
		return model.Comment{
			CreatedAt:   comment.CreateAt(),
			Content:     comment.GetContent(),
			CreatorID:   comment.CreatorID(),
			State:       model.CommentState(comment.GetState()),
			ID:          comment.CommentID(),
			SubComments: make([]model.SubComment, 0, 4),
		}
	})

	return slice.ToMap(parents, func(item model.Comment) model.CommentID {
		return item.ID
	}), nil
}

func (r mysqlRepo) getSubComments(
	ctx context.Context, commentType domain.CommentType, id model.TopicID, ids ...model.CommentID,
) ([]mysqlComment, error) {
	commentIDs := slice.Map(ids, func(item model.CommentID) uint32 {
		return uint32(item)
	})

	var comments []mysqlComment
	var err error
	switch commentType {
	case domain.CommentTypeGroupTopic:
		comments, err = wrapCommentDao(r.q.GroupTopicComment.WithContext(ctx).
			Where(r.q.GroupTopicComment.Related.In(commentIDs...), r.q.GroupTopicComment.TopicID.Eq(uint32(id))).
			Order(r.q.GroupTopicComment.Related, r.q.GroupTopicComment.CreatedTime).Find())
	case domain.CommentTypeSubjectTopic:
		comments, err = wrapCommentDao(r.q.SubjectTopicComment.WithContext(ctx).
			Where(r.q.SubjectTopicComment.Related.In(commentIDs...), r.q.SubjectTopicComment.TopicID.Eq(uint32(id))).
			Order(r.q.SubjectTopicComment.Related, r.q.SubjectTopicComment.CreatedTime).Find())
	case domain.CommentIndex:
		comments, err = wrapCommentDao(r.q.IndexComment.WithContext(ctx).
			Where(r.q.IndexComment.Related.In(commentIDs...), r.q.IndexComment.TopicID.Eq(uint32(id))).
			Order(r.q.IndexComment.Related, r.q.IndexComment.CreatedTime).Find())
	case domain.CommentCharacter:
		comments, err = wrapCommentDao(r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.Related.In(commentIDs...), r.q.CharacterComment.TopicID.Eq(uint32(id))).
			Order(r.q.CharacterComment.Related, r.q.CharacterComment.CreatedTime).Find())
	case domain.CommentPerson:
		comments, err = wrapCommentDao(r.q.PersonComment.WithContext(ctx).
			Where(r.q.PersonComment.Related.In(commentIDs...), r.q.PersonComment.TopicID.Eq(uint32(id))).
			Order(r.q.PersonComment.Related, r.q.PersonComment.CreatedTime).Find())
	case domain.CommentEpisode:
		comments, err = wrapCommentDao(r.q.EpisodeComment.WithContext(ctx).
			Where(r.q.EpisodeComment.Related.In(commentIDs...), r.q.EpisodeComment.TopicID.Eq(uint32(id))).
			Order(r.q.EpisodeComment.Related, r.q.EpisodeComment.CreatedTime).Find())
	default:
		return nil, errUnsupportedCommentType
	}

	if err != nil {
		r.log.Error("failed to get sub replies")
		return nil, errgo.Wrap(err, "dal")
	}

	return comments, nil
}
