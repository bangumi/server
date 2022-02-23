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
	"github.com/bangumi/server/model"
)

type CharacterV0 struct {
	ID        uint32             `json:"id"`
	Name      string             `json:"name"`
	Type      uint8              `json:"type"`
	Images    model.PersonImages `json:"images"`
	Summary   string             `json:"summary"`
	Locked    bool               `json:"locked"`
	Infobox   v0wiki             `json:"infobox"`
	Gender    *string            `json:"gender"`
	BloodType *uint8             `json:"blood_type"`
	BirthYear *uint16            `json:"birth_year"`
	BirthMon  *uint8             `json:"birth_mon"`
	BirthDay  *uint8             `json:"birth_day"`
	Stat      Stat               `json:"stat"`
	Redirect  uint32             `json:"-"` // http 302 response
	NSFW      bool               `json:"nsfw"`
}
