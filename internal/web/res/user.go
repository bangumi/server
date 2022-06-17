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

import (
	"github.com/bangumi/server/internal/model"
)

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

func (a Avatar) Select(s string) (string, bool) {
	switch s {
	case imageSizeLarge:
		return a.Large, true
	case imageSizeMedium:
		return a.Medium, true
	case imageSizeSmall:
		return a.Small, true
	default:
		return "", false
	}
}

type User struct {
	Avatar    Avatar       `json:"avatar"`
	Sign      string       `json:"sign"`
	URL       string       `json:"url"`
	Username  string       `json:"username"`
	Nickname  string       `json:"nickname"`
	ID        model.UserID `json:"id"`
	UserGroup uint8        `json:"user_group"`
}

type Creator struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type LoginRemain struct {
	Remain int `json:"remain"`
}
