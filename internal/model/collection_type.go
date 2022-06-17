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

//go:generate go run golang.org/x/tools/cmd/stringer -type CollectionType -linecomment

type CollectionType uint8

const (
	CollectionWish    CollectionType = 1 // 想看
	CollectionDone    CollectionType = 2 // 看过
	CollectionDoing   CollectionType = 3 // 在看
	CollectionOnHold  CollectionType = 4 // 搁置
	CollectionDropped CollectionType = 5 // 抛弃
)
