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

type PrivateGroupProfile struct {
	CreatedAt     time.Time            `json:"created_at"`
	Name          string               `json:"name"`
	Title         string               `json:"title"`
	Description   string               `json:"description" format:"bbcode"`
	Icon          string               `json:"icon" format:"url"`
	RelatedGroups []string             `json:"related_groups"` // 10个
	NewTopics     []PrivateTopic       `json:"new_topics"`     // 6个
	NewMembers    []PrivateGroupMember `json:"new_members"`    // 10个
	TotalMembers  int64                `json:"total_members"`
}

type PrivateTopic struct {
}

type PrivateGroupMember struct {
	ID       model.UserID `json:"id"`
	UserName string       `json:"user_name"`
	NickName string       `json:"nick_name"`
	Avatar   Avatar       `json:"avatar"`
}
