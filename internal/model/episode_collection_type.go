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

type EpisodeCollectionType uint8

const (
	EpisodeCollectionTypeAll     EpisodeCollectionType = 0 // 全部
	EpisodeCollectionTypeWish    EpisodeCollectionType = 1 // 想看
	EpisodeCollectionTypeDone    EpisodeCollectionType = 2 // 看过
	EpisodeCollectionTypeDropped EpisodeCollectionType = 3 // 抛弃
)
