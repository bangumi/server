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

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
)

var errUnSupportTopicType = errors.New("topic type not support")

func (r mysqlRepo) Get(ctx context.Context, topicType domain.TopicType, id model.TopicID) (model.Topic, error) {
	var topic mysqlTopic
	var err error
	switch topicType {
	case domain.TopicTypeGroup:
		topic, err = r.q.GroupTopic.WithContext(ctx).Where(r.q.GroupTopic.ID.Eq(uint32(id))).First()
	case domain.TopicTypeSubject:
		topic, err = r.q.SubjectTopic.WithContext(ctx).Where(r.q.SubjectTopic.ID.Eq(uint32(id))).First()
	default:
		return model.Topic{}, errUnSupportTopicType
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Topic{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return model.Topic{}, errgo.Wrap(err, "dal")
	}

	return model.Topic{
		CreatedAt: topic.GetCreateTime(),
		UpdatedAt: topic.GetUpdateTime(),
		Title:     topic.GetTitle(),
		ID:        model.TopicID(topic.GetID()),
		CreatorID: model.UserID(topic.GetCreatorID()),
		State:     model.TopicState(topic.GetState()),
		Replies:   topic.GetReplies(),
		ObjectID:  topic.GetObjectID(),
		Status:    model.TopicStatus(topic.GetStatus()),
	}, nil
}

func (r mysqlRepo) Count(
	ctx context.Context, topicType domain.TopicType, id uint32, statuses []model.TopicStatus,
) (int64, error) {
	var (
		count int64
		err   error
	)
	switch topicType {
	case domain.TopicTypeGroup:
		count, err = r.q.GroupTopic.WithContext(ctx).Where(r.q.GroupTopic.GroupID.Eq(id)).
			Where(r.q.GroupTopic.Status.In(slice.ToUint8(statuses)...)).Count()
	case domain.TopicTypeSubject:
		count, err = r.q.SubjectTopic.WithContext(ctx).Where(r.q.SubjectTopic.SubjectID.Eq(id)).
			Where(r.q.SubjectTopic.Status.In(slice.ToUint8(statuses)...)).Count()
	default:
		return 0, errUnSupportTopicType
	}
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}
	return count, nil
}

func (r mysqlRepo) List(
	ctx context.Context, topicType domain.TopicType, id uint32, statuses []model.TopicStatus, limit int, offset int,
) ([]model.Topic, error) {
	var (
		topics []model.Topic
		err    error
	)
	switch topicType {
	case domain.TopicTypeGroup:
		topics, err = wrapDao(r.q.GroupTopic.WithContext(ctx).Where(r.q.GroupTopic.GroupID.Eq(id)).
			Where(r.q.GroupTopic.Status.In(slice.ToUint8(statuses)...)).
			Offset(offset).Limit(limit).Order(r.q.GroupTopic.UpdatedTime.Desc()).Find())
	case domain.TopicTypeSubject:
		topics, err = wrapDao(r.q.SubjectTopic.WithContext(ctx).
			Where(r.q.SubjectTopic.SubjectID.Eq(id)).Where(r.q.SubjectTopic.Status.In(slice.ToUint8(statuses)...)).
			Offset(offset).Limit(limit).Order(r.q.SubjectTopic.UpdatedTime.Desc()).Find())
	default:
		return nil, errUnSupportTopicType
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	return topics, nil
}
