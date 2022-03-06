// Copyright (c) 2021-2022 Trim21 <trim21.me@gmail.com>
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

type SubjectType = uint8

const (
	SubjectBook SubjectType = iota + 1
	SubjectAnime
	SubjectMusic
	SubjectGame
	_
	SubjectReal
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
	case SubjectBook:
		return textSubjectBook
	case SubjectAnime:
		return textSubjectAnime
	case SubjectMusic:
		return textSubjectMusic
	case SubjectGame:
		return textSubjectGame
	case SubjectReal:
		return textSubjectReal
	default:
		return "unknown repository type"
	}
}
