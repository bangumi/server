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

	"github.com/bangumi/server/internal/compat"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/pkg/wiki"
)

type PersonV0 struct {
	LastModified time.Time      `json:"last_modified"`
	BloodType    *uint8         `json:"blood_type"`
	BirthYear    *uint16        `json:"birth_year"`
	BirthDay     *uint8         `json:"birth_day"`
	BirthMon     *uint8         `json:"birth_mon"`
	Gender       *string        `json:"gender"`
	Images       PersonImages   `json:"images"`
	Summary      string         `json:"summary"`
	Name         string         `json:"name"`
	Img          string         `json:"img"`
	Infobox      v0wiki         `json:"infobox"`
	Career       []string       `json:"career"`
	Stat         Stat           `json:"stat"`
	Redirect     model.PersonID `json:"-"`
	ID           model.PersonID `json:"id"`
	Locked       bool           `json:"locked"`
	Type         uint8          `json:"type"`
}

func ConvertModelPerson(s model.Person) PersonV0 {
	img := PersonImage(s.Image)

	return PersonV0{
		ID:           s.ID,
		Type:         s.Type,
		Name:         s.Name,
		Career:       s.Careers(),
		Images:       img,
		Summary:      s.Summary,
		LastModified: time.Time{},
		Infobox:      compat.V0Wiki(wiki.ParseOmitError(s.Infobox).NonZero()),
		Gender:       null.NilString(GenderMap[s.FieldGender]),
		BloodType:    null.NilUint8(s.FieldBloodType),
		BirthYear:    null.NilUint16(s.FieldBirthYear),
		BirthMon:     null.NilUint8(s.FieldBirthMon),
		BirthDay:     null.NilUint8(s.FieldBirthDay),
		Stat: Stat{
			Comments: s.CommentCount,
			Collects: s.CollectCount,
		},
		Img:      img.Large,
		Redirect: s.Redirect,
		Locked:   s.Locked,
	}
}
