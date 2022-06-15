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
)

// User is visible for everyone.
type User struct {
	RegistrationTime time.Time
	NickName         string
	Avatar           string
	Sign             string
	UserName         string
	ID               UserID
	UserGroup        GroupID
}

const (
	// CollectPrivacyNone 默认公开收藏。
	CollectPrivacyNone = 0
	// CollectPrivacySelf 私有收藏，正常计入评分。
	CollectPrivacySelf = 1
	// CollectPrivacyBan Shadow Ban, 显示为私有收藏，不计入评分。
	CollectPrivacyBan = 2
)

type Collection struct {
	UpdatedAt   time.Time
	Comment     string
	Tags        []string
	VolStatus   uint32
	EpStatus    uint32
	SubjectID   SubjectID
	SubjectType uint8
	Rate        uint8
	Type        uint8
	Private     bool
}
