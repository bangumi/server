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
)

type Topics struct {
	Data    []Topic `json:"data"`
	HasMore bool    `json:"has_more"`
	Limit   uint32  `json:"limit"`
	Offset  uint32  `json:"offset"`
}

type Topic struct {
	CreatedAt time.Time `json:"created_at"`
	Comments  *Comments `json:"comments,omitempty"`
	Title     string    `json:"title"`
	Creator   User      `json:"creator"`
	ID        uint32    `json:"id"`
	Replies   uint32    `json:"replies"`
}

type Comments struct {
	Data    []Comment `json:"data"`
	HasMore bool      `json:"has_more"`
	Limit   uint32    `json:"limit"`
	Offset  uint32    `json:"offset"`
}

type Comment struct {
	CreatedAt time.Time `json:"created_at"`
	Creator   User      `json:"creator"`
	Text      string    `json:"text"`
	Replies   []Comment `json:"replies,omitempty"`
	ID        uint32    `json:"id"`
}
