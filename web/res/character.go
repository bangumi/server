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
	wiki "github.com/bangumi/wiki-parser-go"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/compat"
	"github.com/bangumi/server/internal/pkg/null"
)

type CharacterV0 struct {
	BirthMon  *uint8            `json:"birth_mon"`
	Gender    *string           `json:"gender"`
	BirthDay  *uint8            `json:"birth_day"`
	BirthYear *uint16           `json:"birth_year"`
	BloodType *uint8            `json:"blood_type"`
	Images    PersonImages      `json:"images"`
	Summary   string            `json:"summary"`
	Name      string            `json:"name"`
	Infobox   V0wiki            `json:"infobox"`
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

//nolint:gochecknoglobals
var characterStaffMap = map[uint8]string{
	1: "主角",
	2: "配角",
	3: "客串",
	4: "闲角",
	5: "旁白",
	6: "声库",
}

func CharacterStaffString(i uint8) string {
	return characterStaffMap[i]
}

func ConvertModelCharacter(s model.Character) CharacterV0 {
	img := PersonImage(s.Image)

	return CharacterV0{
		ID:        s.ID,
		Type:      s.Type,
		Name:      s.Name,
		NSFW:      s.NSFW,
		Images:    img,
		Summary:   s.Summary,
		Infobox:   compat.V0Wiki(wiki.ParseOmitError(s.Infobox).NonZero()),
		Gender:    null.NilString(GenderMap[s.FieldGender]),
		BloodType: null.NilUint8(s.FieldBloodType),
		BirthYear: null.NilUint16(s.FieldBirthYear),
		BirthMon:  null.NilUint8(s.FieldBirthMon),
		BirthDay:  null.NilUint8(s.FieldBirthDay),
		Stat: Stat{
			Comments: s.CommentCount,
			Collects: s.CollectCount,
		},
		Redirect: s.Redirect,
		Locked:   s.Locked,
	}
}
