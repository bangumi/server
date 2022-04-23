// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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

import "time"

// User is visible for everyone.
type User struct {
	UserName  string
	NickName  string
	Avatar    string
	Sign      string
	ID        uint32
	UserGroup uint8
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
	SubjectType uint8
	HasComment  bool
	Private     bool
	Type        uint8
	VolStatus   uint32
	EpStatus    uint32
	SubjectID   uint32
	Rate        uint8
}
