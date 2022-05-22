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

type PersonV0 struct {
	LastModified time.Time    `json:"last_modified"`
	BloodType    *uint8       `json:"blood_type"`
	BirthYear    *uint16      `json:"birth_year"`
	BirthDay     *uint8       `json:"birth_day"`
	BirthMon     *uint8       `json:"birth_mon"`
	Gender       *string      `json:"gender"`
	Images       PersonImages `json:"images"`
	Summary      string       `json:"summary"`
	Name         string       `json:"name"`
	Img          string       `json:"img"`
	Infobox      v0wiki       `json:"infobox"`
	Career       []string     `json:"career"`
	Stat         Stat         `json:"stat"`
	Redirect     uint32       `json:"-"`
	ID           uint32       `json:"id"`
	Locked       bool         `json:"locked"`
	Type         uint8        `json:"type"`
}
