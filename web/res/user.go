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

package res

type Avatar struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
	Small  string `json:"small"`
}

func (a Avatar) Fill(s string) Avatar {
	if s == "" {
		s = "icon.jpg"
	}
	return Avatar{
		Large:  "https://lain.bgm.tv/pic/user/l/" + s,
		Medium: "https://lain.bgm.tv/pic/user/m/" + s,
		Small:  "https://lain.bgm.tv/pic/user/s/" + s,
	}
}

type Me struct {
	Avatar    Avatar `json:"avatar"`
	Sign      string `json:"sign"`
	URL       string `json:"url"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	ID        uint32 `json:"id"`
	UserGroup uint8  `json:"user_group"`
}
