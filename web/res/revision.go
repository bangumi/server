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

import (
	"time"

	"github.com/bangumi/server/internal/dal/dao"
)

type Revision struct {
	CreatedAt time.Time                 `json:"created_at"`
	Data      dao.GzipPhpSerializedBlob `json:"data"`
	Creator   Creator                   `json:"creator"`
	Summary   string                    `json:"summary"`
	ID        uint32                    `json:"id"`
	Type      uint8                     `json:"type"`
}
