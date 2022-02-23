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

type Character struct {
	Name           string
	Image          string
	Infobox        string
	Summary        string
	ID             uint32
	Redirect       uint32
	CollectCount   uint32
	CommentCount   uint32
	FieldBirthYear uint16
	Producer       bool
	Type           uint8
	Artist         bool
	Seiyu          bool
	Writer         bool
	Illustrator    bool
	Actor          bool
	FieldBloodType uint8
	FieldGender    uint8
	FieldBirthMon  uint8
	Locked         bool
	FieldBirthDay  uint8
	NSFW           bool
}

type PersonCharacterRelation struct {
	Type uint8
}
