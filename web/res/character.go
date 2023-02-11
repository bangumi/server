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

import "github.com/bangumi/server/internal/model"

type CharacterV0 struct {
	BirthMon  *uint8            `json:"birth_mon"`
	Gender    *string           `json:"gender"`
	BirthDay  *uint8            `json:"birth_day"`
	BirthYear *uint16           `json:"birth_year"`
	BloodType *uint8            `json:"blood_type"`
	Images    PersonImages      `json:"images"`
	Summary   string            `json:"summary"`
	Name      string            `json:"name"`
	Infobox   v0wiki            `json:"infobox"`
	Stat      Stat              `json:"stat"`
	ID        model.CharacterID `json:"id"`
	Redirect  model.CharacterID `json:"-"`
	Locked    bool              `json:"locked"`
	Type      uint8             `json:"type"`
	NSFW      bool              `json:"nsfw"`
}

//nolint:gochecknoglobals
var GenderMap = map[uint8]string{
	1: "male",
	2: "female",
}

func CharacterStaffString(i uint8) string {
	switch i {
	case 1:
		return "主角"
	case 2:
		return "配角"
	case 3:
		return "客串"
	}

	return ""
}
