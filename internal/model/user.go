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

import (
	"time"

	"github.com/trim21/go-phpserialize"
)

// User is visible for everyone.
type User struct {
	RegistrationTime time.Time
	NickName         string
	Avatar           string
	Sign             string
	UserName         string
	ID               UserID
	UserGroup        UserGroupID
}

func (u User) GetID() UserID {
	return u.ID
}

type UserSubjectEpisodesCollection map[EpisodeID]UserEpisodeCollection

type UserReceiveFilter uint8

const (
	UserReceiveFilterAll UserReceiveFilter = iota
	UserReceiveFilterFriends
	UserReceiveFilterNone
)

type UserPrivacySettingsField int

const (
	UserPrivacyReceivePrivateMessage      UserPrivacySettingsField = 1
	UserPrivacyReceiveTimelineReply       UserPrivacySettingsField = 30
	UserPrivacyReceiveMentionNotification UserPrivacySettingsField = 20
	UserPrivacyReceiveCommentNotification UserPrivacySettingsField = 21
)

type UserPrivacySettings struct {
	ReceivePrivateMessage      UserReceiveFilter
	ReceiveTimelineReply       UserReceiveFilter
	ReceiveMentionNotification UserReceiveFilter
	ReceiveCommentNotification UserReceiveFilter
}

func (settings *UserPrivacySettings) Unmarshal(s string) {
	rawMap := make(map[int]uint8, 4)
	err := phpserialize.Unmarshal([]byte(s), rawMap)
	if err != nil {
		if v, ok := rawMap[int(UserPrivacyReceivePrivateMessage)]; ok {
			settings.ReceivePrivateMessage = UserReceiveFilter(v)
		}
		if v, ok := rawMap[int(UserPrivacyReceiveTimelineReply)]; ok {
			settings.ReceiveTimelineReply = UserReceiveFilter(v)
		}
		if v, ok := rawMap[int(UserPrivacyReceiveMentionNotification)]; ok {
			settings.ReceiveMentionNotification = UserReceiveFilter(v)
		}
		if v, ok := rawMap[int(UserPrivacyReceiveCommentNotification)]; ok {
			settings.ReceiveCommentNotification = UserReceiveFilter(v)
		}
	}
}

type UserFields struct {
	Site      string
	Location  string
	Bio       string
	Blocklist []UserID
	Privacy   UserPrivacySettings
	UID       UserID
}
