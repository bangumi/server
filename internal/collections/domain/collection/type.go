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

package collection

type SubjectCollection uint8

const (
	SubjectCollectionAll     SubjectCollection = 0 // 全部
	SubjectCollectionWish    SubjectCollection = 1 // 想看
	SubjectCollectionDone    SubjectCollection = 2 // 看过
	SubjectCollectionDoing   SubjectCollection = 3 // 在看
	SubjectCollectionOnHold  SubjectCollection = 4 // 搁置
	SubjectCollectionDropped SubjectCollection = 5 // 抛弃
)

type EpisodeCollection uint8

const (
	EpisodeCollectionNone    EpisodeCollection = 0 // 撤消/删除
	EpisodeCollectionAll     EpisodeCollection = 0 // 全部
	EpisodeCollectionWish    EpisodeCollection = 1 // 想看
	EpisodeCollectionDone    EpisodeCollection = 2 // 看过
	EpisodeCollectionDropped EpisodeCollection = 3 // 抛弃
)
