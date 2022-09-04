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
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/gtime"
)

// ListTopicDisplays 在帖子列表能看到哪些状态的帖子。
func ListTopicDisplays(u domain.Auth) []model.TopicDisplay {
	if u.ID == 0 {
		return []model.TopicDisplay{model.TopicDisplayNormal}
	}

	if u.Permission.ManageTopicState || u.Permission.BanPost {
		return []model.TopicDisplay{model.TopicDisplayBan, model.TopicDisplayNormal, model.TopicDisplayReview}
	}

	return []model.TopicDisplay{model.TopicDisplayNormal}
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

func RewriteCommentTree(comments []model.Comment) []model.Comment {
	var newComments = make([]model.Comment, len(comments))

	for i, comment := range comments {
		comment.SubComments = slice.Map(comment.SubComments, RewriteSubCommit)
		newComments[i] = RewriteCommit(comment)
	}

	return newComments
}

func CanViewTopicContent(u domain.Auth, topic model.Topic) bool {
	if u.ID == 0 {
		// 未登录用户只能看到正常帖子
		return topic.State == model.CommentStateNone && topic.Display == model.TopicDisplayNormal
	}

	// 登录用户

	// 管理员啥都能看
	if u.Permission.ManageTopicState || u.Permission.BanPost {
		return true
	}

	if u.ID == topic.CreatorID {
		if topic.Display == model.TopicDisplayReview {
			return true
		}
	}

	// 非管理员看不到删除和review的帖子
	if topic.Display != model.TopicDisplayNormal {
		return false
	}

	// 注册时间决定
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

const CanViewStateClosedTopic = gtime.OneDay * 180
const CanViewStateDeleteTopic = gtime.OneDay * 365

func CanViewDeleteTopic(a domain.Auth) bool {
	return a.RegisteredLongerThan(CanViewStateDeleteTopic)
}

func CanViewClosedTopic(a domain.Auth) bool {
	return a.RegisteredLongerThan(CanViewStateClosedTopic)
}
