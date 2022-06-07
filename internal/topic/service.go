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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
)

func NewService(s domain.TopicRepo, m domain.CommentRepo) domain.TopicService {
	return service{repo: s, m: m}
}

type service struct {
	repo domain.TopicRepo
	m    domain.CommentRepo
}

func (s service) Get(
	ctx context.Context, topicType domain.TopicType, id model.TopicIDType, limit int, offset int,
) (model.Topic, error) {
	topic, err := s.repo.Get(ctx, topicType, id, limit, offset)
	if err != nil {
		return model.Topic{}, errgo.Wrap(err, "TopicRepo.Get")
	}

	commentType := map[domain.TopicType]domain.CommentType{
		domain.TopicTypeGroup:   domain.CommentTypeGroupTopic,
		domain.TopicTypeSubject: domain.CommentTypeSubjectTopic,
	}[topicType]

	comments, err := s.m.GetComments(ctx, commentType, topic.ID, limit, offset)
	if err != nil {
		return model.Topic{}, errgo.Wrap(err, "CommentRepo.GetCommentsByMentionedID")
	}
	topic.Comments = comments
	return topic, nil
}

func (s service) Count(ctx context.Context, topicType domain.TopicType, id uint32) (int64, error) {
	topics, err := s.repo.Count(ctx, topicType, id)
	if err != nil {
		return 0, errgo.Wrap(err, "TopicRepo.Count")
	}
	return topics, nil
}

func (s service) ListTopics(
	ctx context.Context, topicType domain.TopicType, id uint32, limit int, offset int,
) ([]model.Topic, error) {
	topics, err := s.repo.ListTopics(ctx, topicType, id, limit, offset)
	if err != nil {
		return nil, errgo.Wrap(err, "TopicRepo.GetTopicsByObjectID")
	}
	return topics, nil
}
