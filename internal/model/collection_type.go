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

type SubjectCollectionType uint8

const (
	SubjectCollectionAll     SubjectCollectionType = 0 // 全部
	SubjectCollectionWish    SubjectCollectionType = 1 // 想看
	SubjectCollectionDone    SubjectCollectionType = 2 // 看过
	SubjectCollectionDoing   SubjectCollectionType = 3 // 在看
	SubjectCollectionOnHold  SubjectCollectionType = 4 // 搁置
	SubjectCollectionDropped SubjectCollectionType = 5 // 抛弃
)

type EpisodeCollectionType uint8

const (
	EpisodeCollectionAll     EpisodeCollectionType = 0 // 全部
	EpisodeCollectionWish    EpisodeCollectionType = 1 // 想看
	EpisodeCollectionDone    EpisodeCollectionType = 2 // 看过
	EpisodeCollectionDropped EpisodeCollectionType = 3 // 抛弃
)
