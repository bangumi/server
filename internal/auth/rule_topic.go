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
	"github.com/bangumi/server/internal/pkg/timex"
)

// 目前列表是根据 display(status) 来判断是否能看到标题，state 决定是否能查看内容。

func TopicStatuses(u domain.Auth) []model.TopicStatus {
	if u.ID == 0 {
		return []model.TopicStatus{model.TopicStatusNormal}
	}

	if u.Permission.ManageTopicState || u.Permission.BanPost {
		return []model.TopicStatus{model.TopicStatusBan, model.TopicStatusNormal, model.TopicStatusReview}
	}

	return []model.TopicStatus{model.TopicStatusNormal}
}

func RewriteSubCommit(t model.SubComment) model.SubComment {
	switch t.State {
	case model.CommentStateUserDelete, model.CommentStateAdminDelete:
		t.Content = ""
	default:
		return t
	}

	return t
}

func RewriteCommit(t model.Comment) model.Comment {
	switch t.State {
	case model.CommentStateUserDelete, model.CommentStateAdminDelete:
		t.Content = ""
	default:
		return t
	}

	return t
}

func CanViewTopicContent(u domain.Auth, topic model.Topic) bool {
	if u.ID == 0 {
		return topic.State == model.CommentStateNone
	}

	if u.Permission.ManageTopicState || u.Permission.BanPost {
		return true
	}

	if u.ID == topic.CreatorID {
		return topic.State != model.CommentStateUserDelete
	}

	switch topic.State {
	case model.CommentStateNone, model.CommentStateAdminReopen,
		model.CommentStateAdminMerge, model.CommentStateAdminPin, model.CommentStateAdminSilentTopic:
		return true
	case model.CommentStateAdminCloseTopic:
		return CanViewClosedTopic(u)
	case model.CommentStateUserDelete:
		return CanViewDeleteTopic(u)
	case model.CommentStateAdminDelete:
		return false
	}

	return false
}

const CanViewStateClosedTopic = timex.OneDay * 180
const CanViewStateDeleteTopic = timex.OneDay * 365

func CanViewDeleteTopic(a domain.Auth) bool {
	return a.RegisteredLongerThan(CanViewStateDeleteTopic)
}

func CanViewClosedTopic(a domain.Auth) bool {
	return a.RegisteredLongerThan(CanViewStateClosedTopic)
}
