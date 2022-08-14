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

package ctrl

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/res"
)

var ErrInvalidInput = errors.New("failed")

func (ctl Ctrl) GetTopic(
	ctx context.Context, u domain.Auth, topicType domain.TopicType, topicID model.TopicID, limit int, offset int,
) (model.TopicDetail, error) {
	var commentType domain.CommentType
	switch topicType {
	case domain.TopicTypeGroup:
		commentType = domain.CommentTypeGroupTopic
	case domain.TopicTypeSubject:
		commentType = domain.CommentTypeSubjectTopic
	default:
		return model.TopicDetail{}, ErrInvalidInput
	}

	topic, err := ctl.topic.Get(ctx, topicType, topicID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return model.TopicDetail{}, domain.ErrNotFound
		}

		ctl.log.Error("failed to get topic", zap.Error(err), log.TopicType(topicType), log.TopicID(topicID))
		return model.TopicDetail{}, errgo.Wrap(err, "topicRepo.Get")
	}

	if !auth.CanViewTopicContent(u, topic) {
		return model.TopicDetail{}, domain.ErrNotFound
	}

	content, err := ctl.topic.GetTopicContent(ctx, topicType, topicID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return model.TopicDetail{}, res.ErrNotFound
		}

		ctl.log.Error("failed to get topic content", zap.Error(err), log.TopicType(topicType), log.TopicID(topicID))
		return model.TopicDetail{}, errgo.Wrap(err, "topic.GetTopicContent")
	}

	comments, err := ctl.topic.ListReplies(ctx, commentType, topic.ID, limit, offset)
	if err != nil {
		return model.TopicDetail{}, errgo.Wrap(err, "topicRepo.ListReplies")
	}

	return model.TopicDetail{
		Topic:   topic,
		Content: auth.RewriteCommit(content).Content,
		Replies: auth.RewriteCommentTree(comments),
	}, nil
}

func (ctl Ctrl) ListTopics(
	ctx context.Context,
	u domain.Auth,
	topicType domain.TopicType,
	objectID uint32,
	limit, offset int,
) ([]model.Topic, int64, error) {
	statuses := auth.TopicStatuses(u)

	count, err := ctl.topic.Count(ctx, topicType, objectID, statuses)
	if err != nil {
		return nil, 0, errgo.Wrap(err, "topicRepo.Count")
	}

	if count == 0 || int64(offset) > count {
		return []model.Topic{}, 0, nil
	}

	topics, err := ctl.topic.List(ctx, topicType, objectID, statuses, limit, offset)
	if err != nil {
		return nil, 0, errgo.Wrap(err, "repo.topic.GetTopics")
	}

	return topics, count, nil
}

func (ctl Ctrl) ListReplies(
	ctx context.Context,
	commentType domain.CommentType,
	topicID model.TopicID,
	limit, offset int,
) ([]model.Comment, int64, error) {
	count, err := ctl.topic.CountReplies(ctx, commentType, topicID)
	if err != nil {
		return nil, 0, errgo.Wrap(err, "topicRepo.Count")
	}

	if count == 0 || int64(offset) > count {
		return nil, 0, nil
	}

	comments, err := ctl.topic.ListReplies(ctx, commentType, topicID, limit, offset)
	if err != nil {
		return nil, 0, errgo.Wrap(err, "repo.topic.GetTopics")
	}

	return auth.RewriteCommentTree(comments), count, nil
}
