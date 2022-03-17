// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
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
)

type PersonRevisionDataItem struct {
	InfoBox    string            `json:"prsn_infobox"`
	Summary    string            `json:"prsn_summary"`
	Profession map[string]string `json:"profession"`
	Extra      map[string]string `json:"extra"`
	Name       string            `json:"prsn_name"`
}

type PersonRevision struct {
	CreatedAt time.Time                         `json:"created_at"`
	Data      map[string]PersonRevisionDataItem `json:"data"`
	Creator   Creator                           `json:"creator"`
	Summary   string                            `json:"summary"`
	ID        uint32                            `json:"id"`
	Type      uint8                             `json:"type"`
}
