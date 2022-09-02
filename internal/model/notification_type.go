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
	NotificationTypeGroupTopicReply          NotificationType = iota + 1 // 发起的小组话题有新回复
	NotificationTypeReplyToGroupTopicReply                               // 在小组话题收到回复
	NotificationTypeSubjectTopicReply                                    // 发起的条目讨论有新回复
	NotificationTypeReplyToSubjectTopicReply                             // 在条目讨论收到回复
	NotificationTypeCharacterMessage                                     // 关注的角色讨论有新回复
	NotificationTypeReplyToCharacterMessage                              // 在角色讨论收到回复
	NotificationTypeBlogMessage                                          // 日志留言
	NotificationTypeReplyToBlogMessage                                   // 日志留言的回复
	NotificationTypeEpisodeTopicReply                                    // 章节讨论有新回复
	NotificationTypeReplyToEpisodeTopic                                  // 在章节讨论收到回复
	NotificationTypeIndexMessage                                         // 目录有新留言
	NotificationTypeReplyToIndexMessage                                  // 目录留言收到回复
	NotificationTypeReplyToPersonMessage                                 // 人物留言收到回复
	NotificationTypeFriendRequest                                        // 收到好友申请
	NotificationTypePassFriendRequest                                    // 好友申请通过
	_
	NotificationTypeDoujinClubTopicReply           //    同人社团讨论有新回复
	NotificationTypeReplyToDoujinClubTopicReply    //    在同人社团讨论收到回复
	NotificationTypeReplyToDoujinSubjectTopicReply //    同人作品讨论有新回复
	NotificationTypeDoujinEventTopicReply          //    同人展会讨论有新回复
	NotificationTypeReplyToDoujinEventTopicReply   //    在同人展会讨论收到回复
	NotificationTypeTsukkomiReply                  //    吐槽有新回复
	NotificationTypeGroupTopicMention              //    在小组讨论中被提及
	NotificationTypeSubjectTopicMention            //    在条目讨论中被提及
	NotificationTypeCharacterMessageMention        //    在角色留言中被提及
	NotificationTypePersonMessageMention           //    在人物留言中被提及
	NotificationTypeIndexMessageMention            //    在目录留言中被提及
	NotificationTypeTukkomiMention                 //    在吐槽中被提及
	NotificationTypeBlogMessageMention             //    在日志留言中被提及
	NotificationTypeEpisodeTopicMention            //    在章节讨论中被提及
	NotificationTypeDoujinClubMessageMention       //    在同人社团留言中被提及
	NotificationTypeDoujinClubTopicMention         //    在同人社团讨论中被提及
	NotificationTypeDoujinSubjectMessageMention    //    在同人作品留言中被提及
	NotificationTypeDoujinEventTopicMention        //    在同人展会讨论中被提及
)

type NotificationStatus uint8

const (
	NotificationStatusRead NotificationStatus = iota
	NotificationStatusUnread
)
