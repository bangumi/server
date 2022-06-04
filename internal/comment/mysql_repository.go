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
	ctx context.Context, commentType model.CommentType, id model.CommentIDType,
) (model.Comment, error) {
	var (
		comment interface{}
		err     error
	)
	switch commentType {
	case model.CommentTypeGroupTopic:
		comment, err = r.q.GroupTopicComment.WithContext(ctx).Where(r.q.GroupTopicComment.ID.Eq(id)).First()
	case model.CommentTypeSubjectTopic:
		comment, err = r.q.SubjectTopicComment.WithContext(ctx).Where(r.q.SubjectTopicComment.ID.Eq(id)).First()
	case model.CommentIndex:
		comment, err = r.q.IndexComment.WithContext(ctx).Where(r.q.IndexComment.ID.Eq(id)).First()
	case model.CommentCharacter:
		comment, err = r.q.CharacterComment.WithContext(ctx).Where(r.q.CharacterComment.ID.Eq(id)).First()
	case model.CommentEpisode:
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

	return ConvertDao(comment)
}

var errUnsupportCommentType = errors.New("comment type not support")

func ConvertDao(in interface{}) (model.Comment, error) {
	switch v := in.(type) {
	case *dao.SubjectTopicComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         v.UID,
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.Dateline), 0),
			Content:     v.Content,
		}, nil
	case *dao.GroupTopicComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         v.UID,
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.Dateline), 0),
			Content:     v.Content,
		}, nil
	case *dao.IndexComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         v.UID,
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.Dateline), 0),
			Content:     v.Content,
		}, nil
	case *dao.EpisodeComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         v.UID,
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.Dateline), 0),
			Content:     v.Content,
		}, nil
	case *dao.CharacterComment:
		return model.Comment{
			ID:          v.ID,
			MentionedID: v.MentionedID,
			UID:         v.UID,
			Related:     v.Related,
			CreatedAt:   time.Unix(int64(v.Dateline), 0),
			Content:     v.Content,
		}, nil
	default:
		return model.Comment{}, errUnsupportCommentType
	}
}

func (r mysqlRepo) GetCommentsByMentionedID(
	ctx context.Context, commentType model.CommentType, limit int, offset int, id uint32,
) (model.Comments, error) {
	var (
		comments interface{}
		total    int64
		err      error
	)
	switch commentType {
	case model.CommentTypeGroupTopic:
		comments, total, err = r.q.GroupTopicComment.WithContext(ctx).
			Where(r.q.GroupTopicComment.MentionedID.Eq(id)).FindByPage(offset, limit)
	case model.CommentTypeSubjectTopic:
		comments, total, err = r.q.SubjectTopicComment.WithContext(ctx).
			Where(r.q.SubjectTopicComment.MentionedID.Eq(id)).FindByPage(offset, limit)
	case model.CommentIndex:
		comments, total, err = r.q.IndexComment.WithContext(ctx).
			Where(r.q.IndexComment.MentionedID.Eq(id)).FindByPage(offset, limit)
	case model.CommentCharacter:
		comments, total, err = r.q.CharacterComment.WithContext(ctx).
			Where(r.q.CharacterComment.MentionedID.Eq(id)).FindByPage(offset, limit)
	case model.CommentEpisode:
		comments, total, err = r.q.EpisodeComment.WithContext(ctx).
			Where(r.q.EpisodeComment.MentionedID.Eq(id)).FindByPage(offset, limit)
	default:
		return model.Comments{}, errUnsupportCommentType
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Comments{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return model.Comments{}, errgo.Wrap(err, "dal")
	}

	return model.Comments{
		Total:  uint32(total),
		Limit:  uint32(limit),
		Offset: uint32(offset),
		Data:   convertModelComments(comments),
	}, nil
}

func convertModelComments(in interface{}) []model.Comment {
	comments := make([]model.Comment, 0)
	switch list := in.(type) {
	case []*dao.SubjectTopicComment:
		for _, v := range list {
			if comment, e := ConvertDao(v); e == nil {
				comments = append(comments, comment)
			}
		}
	case []*dao.GroupTopicComment:
		for _, v := range list {
			if comment, e := ConvertDao(v); e == nil {
				comments = append(comments, comment)
			}
		}
	}
	return comments
}
