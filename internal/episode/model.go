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

package episode

import (
	"github.com/bangumi/server/internal/model"
)

type Type = uint8

const (
	TypeNormal  Type = 0
	TypeSpecial Type = 1
	TypeOpening Type = 2
	TypeEnding  Type = 3
	TypeMad     Type = 4
	TypeOther   Type = 6
)

type Episode struct {
	Airdate     string
	Name        string
	NameCN      string
	Duration    string
	Description string
	Ep          float32
	SubjectID   model.SubjectID
	Sort        float32
	Comment     uint32
	ID          model.EpisodeID
	Type        Type
	Disc        uint8
}

func (e Episode) Less(o Episode) bool {
	if e.Type == o.Type {
		return e.Sort <= o.Sort
	}

	return e.Type <= o.Type
}
