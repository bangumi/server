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

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.TopicRepo, error) {
	return mysqlRepo{q: q, log: log.Named("subject.mysqlRepo")}, nil
}

func (r mysqlRepo) Get(ctx context.Context, topicType domain.TopicType, id model.TopicIDType) (model.Topic, error) {
	var (
		topic interface{}
		err   error
	)
	switch topicType {
	case domain.TopicTypeGroup:
		topic, err = r.q.GroupTopic.WithContext(ctx).Where(r.q.GroupTopic.ID.Eq(id)).First()
	case domain.TopicTypeSubject:
		topic, err = r.q.SubjectTopic.WithContext(ctx).Where(r.q.SubjectTopic.ID.Eq(id)).First()
	default:
		return model.Topic{}, errUnsupportTopicType
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Topic{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return model.Topic{}, errgo.Wrap(err, "dal")
	}

	return convertDao(topic)
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
			Where(r.q.GroupTopic.Status.In(convertToUint8Status(statuses)...)).Count()
	case domain.TopicTypeSubject:
		count, err = r.q.SubjectTopic.WithContext(ctx).Where(r.q.SubjectTopic.SubjectID.Eq(id)).
			Where(r.q.SubjectTopic.Status.In(convertToUint8Status(statuses)...)).Count()
	default:
		return 0, errUnsupportTopicType
	}
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}
	return count, nil
}

func convertToUint8Status(statuses []model.TopicStatus) []uint8 {
	s := make([]uint8, len(statuses))
	for i, v := range statuses {
		s[i] = uint8(v)
	}
	return s
}

func (r mysqlRepo) ListTopics(
	ctx context.Context, topicType domain.TopicType, id uint32, statuses []model.TopicStatus, limit int, offset int,
) ([]model.Topic, error) {
	var (
		topics interface{}
		err    error
	)
	switch topicType {
	case domain.TopicTypeGroup:
		topics, err = r.q.GroupTopic.WithContext(ctx).Where(r.q.GroupTopic.GroupID.Eq(id)).
			Where(r.q.GroupTopic.Status.In(convertToUint8Status(statuses)...)).
			Offset(offset).Limit(limit).Find()
	case domain.TopicTypeSubject:
		topics, err = r.q.SubjectTopic.WithContext(ctx).
			Where(r.q.SubjectTopic.SubjectID.Eq(id)).Where(r.q.SubjectTopic.Status.In(convertToUint8Status(statuses)...)).
			Offset(offset).Limit(limit).Find()
	default:
		return nil, errUnsupportTopicType
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}
	return convertModelTopics(topics), nil
}

var errUnsupportTopicType = errors.New("topic type not support")

func convertModelTopics(in interface{}) []model.Topic {
	topics := make([]model.Topic, 0)
	switch list := in.(type) {
	case []*dao.SubjectTopic:
		for _, v := range list {
			if topic, e := convertDao(v); e == nil {
				topics = append(topics, topic)
			}
		}
	case []*dao.GroupTopic:
		for _, v := range list {
			if topic, e := convertDao(v); e == nil {
				topics = append(topics, topic)
			}
		}
	}
	return topics
}

func convertDao(in interface{}) (model.Topic, error) {
	switch v := in.(type) {
	case *dao.GroupTopic:
		return model.Topic{
			ID:        v.ID,
			ObjectID:  v.GroupID,
			UID:       v.UID,
			Title:     v.Title,
			CreatedAt: time.Unix(int64(v.CreatedAt), 0),
			UpdatedAt: time.Unix(int64(v.UpdatedAt), 0),
			Replies:   v.Replies,
			State:     model.TopicState(v.State),
			Status:    model.TopicStatus(v.Status),
		}, nil
	case *dao.SubjectTopic:
		return model.Topic{
			ID:        v.ID,
			ObjectID:  v.SubjectID,
			UID:       v.UID,
			Title:     v.Title,
			CreatedAt: time.Unix(int64(v.CreatedAt), 0),
			UpdatedAt: time.Unix(int64(v.UpdatedAt), 0),
			Replies:   v.Replies,
			State:     model.TopicState(v.State),
			Status:    model.TopicStatus(v.Status),
		}, nil
	default:
		return model.Topic{}, errUnsupportTopicType
	}
}
