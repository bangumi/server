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

package auth

import (
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
)

func TopicStatuses(u domain.Auth) []model.TopicStatus {
	if u.ID == 0 {
		return []model.TopicStatus{model.TopicStatusNormal}
	}

	if u.Permission.ManageTopicState {
		return []model.TopicStatus{model.TopicStatusBan, model.TopicStatusNormal, model.TopicStatusReview}
	}

	var s = make([]model.TopicStatus, 0, 2)
	s = append(s, model.TopicStatusNormal)

	if u.Permission.BanPost {
		s = append(s, model.TopicStatusReview)
	}

	return s
}

func CanViewTopic(u domain.Auth, topic model.Topic) bool {
	if u.ID == 0 {
		return topic.State == model.TopicStateNone
	}

	if u.Permission.ManageTopicState {
		return true
	}

	if u.ID == topic.CreatorID {
		return topic.State != model.TopicStateDelete
	}

	switch topic.State {
	case model.TopicStateNone, model.TopicStateReopen, model.TopicStateMerge, model.TopicStatePin, model.TopicStateSilent:
		return true
	case model.TopicStateClosed:
		return CanViewClosedTopic(u)
	case model.TopicStateDelete:
		return CanViewDeleteTopic(u)
	case model.TopicStatePrivate:
		return false
	}

	return false
}

func CanViewDeleteTopic(a domain.Auth) bool {
	return a.RegisteredLongerThan(CanViewStateDeleteTopic)
}

func CanViewClosedTopic(a domain.Auth) bool {
	return a.RegisteredLongerThan(CanViewStateClosedTopic)
}
