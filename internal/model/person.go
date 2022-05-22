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

package model

type Person struct {
	Name           string
	Image          string
	Infobox        string
	Summary        string
	ID             uint32
	Redirect       uint32
	CollectCount   uint32
	CommentCount   uint32
	FieldBirthYear uint16
	Producer       bool
	Mangaka        bool
	Type           uint8
	Artist         bool
	Seiyu          bool
	Writer         bool
	Illustrator    bool
	Actor          bool
	FieldBloodType uint8
	FieldGender    uint8
	FieldBirthMon  uint8
	Locked         bool
	FieldBirthDay  uint8
}

func (p Person) Careers() []string {
	s := make([]string, 0, 7)

	if p.Writer {
		s = append(s, "writer")
	}

	if p.Producer {
		s = append(s, "producer")
	}

	if p.Mangaka {
		s = append(s, "mangaka")
	}

	if p.Artist {
		s = append(s, "artist")
	}

	if p.Seiyu {
		s = append(s, "seiyu")
	}

	if p.Writer {
		s = append(s, "writer")
	}

	if p.Illustrator {
		s = append(s, "illustrator")
	}

	if p.Actor {
		s = append(s, "actor")
	}

	return s
}
