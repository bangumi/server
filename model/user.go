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
