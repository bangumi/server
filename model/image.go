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

type PersonImages struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
}

type SubjectImages struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
	Common string `json:"common"`
}

func SubjectImage(s string) SubjectImages {
	if s == "" {
		return SubjectImages{}
	}

	return SubjectImages{
		Common: "https://lain.bgm.tv/pic/cover/c/" + s,
		Large:  "https://lain.bgm.tv/pic/cover/l/" + s,
		Medium: "https://lain.bgm.tv/pic/cover/m/" + s,
	}
}
