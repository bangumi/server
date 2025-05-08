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

package user

import (
	"time"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/serialize"
)

// FullUser is for current user or admin only.
type FullUser struct {
	RegistrationTime time.Time
	NickName         string
	Avatar           string
	Sign             string
	UserName         string
	ID               model.UserID
	UserGroup        GroupID
	TimeOffset       int8
	Email            string
}

type GroupID = uint8

// User is visible for everyone.
type User struct {
	RegistrationTime time.Time
	NickName         string
	Avatar           string
	Sign             string
	UserName         string
	ID               model.UserID
	UserGroup        GroupID
}

func (u User) GetID() model.UserID {
	return u.ID
}

type ReceiveFilter uint8

const (
	ReceiveFilterAll ReceiveFilter = iota
	ReceiveFilterFriends
	ReceiveFilterNone
)

type PrivacySettingsField int

const (
	PrivacyReceivePrivateMessage      PrivacySettingsField = 1
	PrivacyReceiveTimelineReply       PrivacySettingsField = 30
	PrivacyReceiveMentionNotification PrivacySettingsField = 20
	PrivacyReceiveCommentNotification PrivacySettingsField = 21
)

type PrivacySettings struct {
	ReceivePrivateMessage      ReceiveFilter
	ReceiveTimelineReply       ReceiveFilter
	ReceiveMentionNotification ReceiveFilter
	ReceiveCommentNotification ReceiveFilter
}

func (settings *PrivacySettings) Unmarshal(s []byte) {
	rawMap := make(map[PrivacySettingsField]ReceiveFilter, 4)
	if len(s) != 0 {
		err := serialize.Decode(s, &rawMap)
		if err != nil {
			return
		}
	}

	settings.ReceivePrivateMessage = rawMap[PrivacyReceivePrivateMessage]
	settings.ReceiveTimelineReply = rawMap[PrivacyReceiveTimelineReply]
	settings.ReceiveMentionNotification = rawMap[PrivacyReceiveMentionNotification]
	settings.ReceiveCommentNotification = rawMap[PrivacyReceiveCommentNotification]
}

type Fields struct {
	Site      string
	Location  string
	Bio       string
	BlockList []model.UserID
	Privacy   PrivacySettings
	UID       model.UserID
}
