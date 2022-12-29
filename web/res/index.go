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
	"github.com/bangumi/server/internal/user"
)

type Index struct {
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Creator     Creator       `json:"creator"`
	Title       string        `json:"title"`
	Description string        `json:"desc"`
	Total       uint32        `json:"total"`
	ID          model.IndexID `json:"id"`
	Stat        Stat          `json:"stat"`
	Ban         bool          `json:"ban"`
	NSFW        bool          `json:"nsfw" doc:"if index contains any nsfw subjects"`
}

func IndexModelToResponse(i *model.Index, u user.User) Index {
	return Index{
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Creator: Creator{
			Username: u.UserName,
			Nickname: u.NickName,
		},
		Title:       i.Title,
		Description: i.Description,
		Total:       i.Total,
		ID:          i.ID,
		Stat: Stat{
			Comments: i.Comments,
			Collects: i.Collects,
		},
		Ban:  i.Ban,
		NSFW: i.NSFW,
	}
}
