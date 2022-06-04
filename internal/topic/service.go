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
	ctx context.Context, topicType model.TopicType, limit int, offset int, id model.TopicIDType,
) (model.Topic, error) {
	topic, err := s.repo.Get(ctx, topicType, limit, offset, id)

	commentType := map[model.TopicType]model.CommentType{
		model.TopicTypeGroup:   model.CommentTypeGroupTopic,
		model.TopicTypeSubject: model.CommentTypeSubjectTopic,
	}[topicType]

	comments, err := s.m.GetCommentsByMentionedID(ctx, commentType, limit, offset, topic.ID)
	if err != nil {
		return model.Topic{}, err
	}
	topic.Comments = comments
	return topic, nil
}

func (s service) GetTopicsByObjectID(ctx context.Context, topicType model.TopicType, id uint32) ([]model.Topic, error) {
	return s.repo.GetTopicsByObjectID(ctx, topicType, id)
}
