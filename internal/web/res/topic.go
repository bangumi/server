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
	UpdatedAt time.Time `json:"updated_at"`
	Comments  *Paged    `json:"comments,omitempty"`
	Title     string    `json:"title"`
	Creator   User      `json:"creator"`
	ID        uint32    `json:"id"`
	Replies   uint32    `json:"replies"`
}

type Comment struct {
	CreatedAt time.Time `json:"created_at"`
	ReplyTo   *Comment  `json:"reply_to,omitempty"`
	Text      string    `json:"text"`
	Creator   User      `json:"creator"`
	ID        uint32    `json:"id"`
}
