// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
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

import (
	"time"
)

type Creator struct {
	Username string
	Nickname string
}

type Revision struct {
	Data      interface{}
	CreatedAt time.Time
	Summary   string
	Type      uint8
	ID        uint32
	CreatorID uint32
}

type SubjectRevisionData struct {
	Name         string
	NameCN       string
	VoteField    string
	FieldInfobox string
	FieldSummary string
	Platform     uint16
	TypeID       uint16
	SubjectID    uint32
	FieldEps     uint32
	Type         uint8
}
