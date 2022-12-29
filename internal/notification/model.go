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

package notification

import (
	"time"

	"github.com/bangumi/server/internal/model"
)

type Field struct {
	Title       string
	ID          model.NotificationFieldID
	RelatedID   uint32 // 关联的主体事项的id，条目讨论，日志的
	RelatedType uint8  // 关联的主体事项的分类
}

type Notification struct {
	CreatedTime time.Time
	ID          model.NotificationID
	ReceiverID  model.UserID
	SenderID    model.UserID
	FieldID     model.NotificationFieldID
	RelatedID   uint32 // 触发通知的实际事项id，如回复的
	Type        Type
	Status      Status
}

type Type uint8

const (
	TypeGroupTopicReply                Type = 1  // 发起的小组话题有新回复
	TypeReplyToGroupTopicReply         Type = 2  // 在小组话题收到回复
	TypeSubjectTopicReply              Type = 3  // 发起的条目讨论有新回复
	TypeReplyToSubjectTopicReply       Type = 4  // 在条目讨论收到回复
	TypeCharacterMessage               Type = 5  // 关注的角色讨论有新回复
	TypeReplyToCharacterMessage        Type = 6  // 在角色讨论收到回复
	TypeBlogMessage                    Type = 7  // 日志留言
	TypeReplyToBlogMessage             Type = 8  // 日志留言的回复
	TypeEpisodeTopicReply              Type = 9  // 章节讨论有新回复
	TypeReplyToEpisodeTopic            Type = 10 // 在章节讨论收到回复
	TypeIndexMessage                   Type = 11 // 目录有新留言
	TypeReplyToIndexMessage            Type = 12 // 目录留言收到回复
	TypeReplyToPersonMessage           Type = 13 // 人物留言收到回复
	TypeFriendRequest                  Type = 14 // 收到好友申请
	TypePassFriendRequest              Type = 15 // 好友申请通过
	TypeDoujinClubTopicReply           Type = 17 // 同人社团讨论有新回复
	TypeReplyToDoujinClubTopicReply    Type = 18 // 在同人社团讨论收到回复
	TypeReplyToDoujinSubjectTopicReply Type = 19 // 同人作品讨论有新回复
	TypeDoujinEventTopicReply          Type = 20 // 同人展会讨论有新回复
	TypeReplyToDoujinEventTopicReply   Type = 21 // 在同人展会讨论收到回复
	TypeTsukkomiReply                  Type = 22 // 吐槽有新回复
	TypeGroupTopicMention              Type = 23 // 在小组讨论中被提及
	TypeSubjectTopicMention            Type = 24 // 在条目讨论中被提及
	TypeCharacterMessageMention        Type = 25 // 在角色留言中被提及
	TypePersonMessageMention           Type = 26 // 在人物留言中被提及
	TypeIndexMessageMention            Type = 27 // 在目录留言中被提及
	TypeTsukkomiMention                Type = 28 // 在吐槽中被提及
	TypeBlogMessageMention             Type = 29 // 在日志留言中被提及
	TypeEpisodeTopicMention            Type = 30 // 在章节讨论中被提及
	TypeDoujinClubMessageMention       Type = 31 // 在同人社团留言中被提及
	TypeDoujinClubTopicMention         Type = 32 // 在同人社团讨论中被提及
	TypeDoujinSubjectMessageMention    Type = 33 // 在同人作品留言中被提及
	TypeDoujinEventTopicMention        Type = 34 // 在同人展会讨论中被提及
)

type Status uint8

const (
	StatusRead   Status = 0
	StatusUnread Status = 1
)
