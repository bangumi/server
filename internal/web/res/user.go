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
	"time"

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

type User struct {
	Avatar    Avatar `json:"avatar"`
	Sign      string `json:"sign"`
	URL       string `json:"url"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	ID        uint32 `json:"id"`
	UserGroup uint8  `json:"user_group"`
}

type Collection struct {
	UpdatedAt   time.Time           `json:"updated_at"`
	Comment     *string             `json:"comment"`
	Tags        []string            `json:"tags"`
	SubjectID   model.SubjectIDType `json:"subject_id"`
	EpStatus    uint32              `json:"ep_status"`
	VolStatus   uint32              `json:"vol_status"`
	SubjectType uint8               `json:"subject_type"`
	Type        uint8               `json:"type"`
	Rate        uint8               `json:"rate"`
	Private     bool                `json:"private"`
}

type Creator struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type LoginRemain struct {
	Remain int `json:"remain"`
}
