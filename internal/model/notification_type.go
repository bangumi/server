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
package model

type NotificationType uint8

const (
	NotificationTypeGroupTopicReply                NotificationType = 1  // 发起的小组话题有新回复
	NotificationTypeReplyToGroupTopicReply         NotificationType = 2  // 在小组话题收到回复
	NotificationTypeSubjectTopicReply              NotificationType = 3  // 发起的条目讨论有新回复
	NotificationTypeReplyToSubjectTopicReply       NotificationType = 4  // 在条目讨论收到回复
	NotificationTypeCharacterMessage               NotificationType = 5  // 关注的角色讨论有新回复
	NotificationTypeReplyToCharacterMessage        NotificationType = 6  // 在角色讨论收到回复
	NotificationTypeBlogMessage                    NotificationType = 7  // 日志留言
	NotificationTypeReplyToBlogMessage             NotificationType = 8  // 日志留言的回复
	NotificationTypeEpisodeTopicReply              NotificationType = 9  // 章节讨论有新回复
	NotificationTypeReplyToEpisodeTopic            NotificationType = 10 // 在章节讨论收到回复
	NotificationTypeIndexMessage                   NotificationType = 11 // 目录有新留言
	NotificationTypeReplyToIndexMessage            NotificationType = 12 // 目录留言收到回复
	NotificationTypeReplyToPersonMessage           NotificationType = 13 // 人物留言收到回复
	NotificationTypeFriendRequest                  NotificationType = 14 // 收到好友申请
	NotificationTypePassFriendRequest              NotificationType = 15 // 好友申请通过
	NotificationTypeDoujinClubTopicReply           NotificationType = 17 //    同人社团讨论有新回复
	NotificationTypeReplyToDoujinClubTopicReply    NotificationType = 18 //    在同人社团讨论收到回复
	NotificationTypeReplyToDoujinSubjectTopicReply NotificationType = 19 //    同人作品讨论有新回复
	NotificationTypeDoujinEventTopicReply          NotificationType = 20 //    同人展会讨论有新回复
	NotificationTypeReplyToDoujinEventTopicReply   NotificationType = 21 //    在同人展会讨论收到回复
	NotificationTypeTsukkomiReply                  NotificationType = 22 //    吐槽有新回复
	NotificationTypeGroupTopicMention              NotificationType = 23 //    在小组讨论中被提及
	NotificationTypeSubjectTopicMention            NotificationType = 24 //    在条目讨论中被提及
	NotificationTypeCharacterMessageMention        NotificationType = 25 //    在角色留言中被提及
	NotificationTypePersonMessageMention           NotificationType = 26 //    在人物留言中被提及
	NotificationTypeIndexMessageMention            NotificationType = 27 //    在目录留言中被提及
	NotificationTypeTukkomiMention                 NotificationType = 28 //    在吐槽中被提及
	NotificationTypeBlogMessageMention             NotificationType = 29 //    在日志留言中被提及
	NotificationTypeEpisodeTopicMention            NotificationType = 30 //    在章节讨论中被提及
	NotificationTypeDoujinClubMessageMention       NotificationType = 31 //    在同人社团留言中被提及
	NotificationTypeDoujinClubTopicMention         NotificationType = 32 //    在同人社团讨论中被提及
	NotificationTypeDoujinSubjectMessageMention    NotificationType = 33 //    在同人作品留言中被提及
	NotificationTypeDoujinEventTopicMention        NotificationType = 34 //    在同人展会讨论中被提及
)

type NotificationStatus uint8

const (
	NotificationStatusRead   NotificationStatus = 0
	NotificationStatusUnread NotificationStatus = 1
)
