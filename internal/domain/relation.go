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

package domain

import (
	"github.com/bangumi/server/internal/model"
)

type PersonCharacterRelation struct {
	CharacterID model.CharacterID
	PersonID    model.PersonID
	SubjectID   model.SubjectID
}

type SubjectPersonRelation struct {
	TypeID uint16

	PersonID  model.PersonID
	SubjectID model.SubjectID
}

type SubjectCharacterRelation struct {
	TypeID uint8

	SubjectID   model.SubjectID
	CharacterID model.CharacterID
}

type SubjectInternalRelation struct {
	TypeID uint16

	SourceID      model.SubjectID
	DestinationID model.SubjectID
}

func (s SubjectInternalRelation) GetSourceID() model.SubjectID {
	return s.SourceID
}
