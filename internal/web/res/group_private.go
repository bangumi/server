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

type PrivateGroup struct {
	ID        model.GroupID `json:"id"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"created_at"`
	Title     string        `json:"title"`
	Icon      string        `json:"icon" format:"url"`
}

type PrivateGroupProfile struct {
	CreatedAt    time.Time            `json:"created_at"`
	Name         string               `json:"name"`
	Title        string               `json:"title"`
	Description  string               `json:"description" format:"bbcode"`
	Icon         string               `json:"icon" format:"url"`
	NewMembers   []PrivateGroupMember `json:"new_members"`
	TotalMembers int64                `json:"total_members"`
	ID           model.GroupID        `json:"id"`
}

type PrivateGroupMember struct {
	JoinedAt time.Time    `json:"joined_at"`
	Avatar   Avatar       `json:"avatar"`
	UserName string       `json:"username"`
	NickName string       `json:"nickname"`
	ID       model.UserID `json:"id"`
}
