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

type SubjectType = uint8

const (
	SubjectTypeAll   SubjectType = 0
	SubjectTypeBook  SubjectType = 1 // 书籍
	SubjectTypeAnime SubjectType = 2 // 动画
	SubjectTypeMusic SubjectType = 3 // 音乐
	SubjectTypeGame  SubjectType = 4 // 游戏
	SubjectTypeReal  SubjectType = 6 // 三次元
)

const (
	textSubjectBook  = "书籍"
	textSubjectAnime = "动画"
	textSubjectMusic = "音乐"
	textSubjectGame  = "游戏"
	textSubjectReal  = "三次元"
)

func SubjectTypeString(s uint8) string {
	switch s {
	case SubjectTypeBook:
		return textSubjectBook
	case SubjectTypeAnime:
		return textSubjectAnime
	case SubjectTypeMusic:
		return textSubjectMusic
	case SubjectTypeGame:
		return textSubjectGame
	case SubjectTypeReal:
		return textSubjectReal
	default:
		return "unknown repository type"
	}
}
