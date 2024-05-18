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

	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
)

type SubjectCollection struct {
	UpdatedAt   time.Time                    `json:"updated_at"`
	Comment     *string                      `json:"comment"`
	Tags        []string                     `json:"tags"`
	Subject     SlimSubjectV0                `json:"subject"`
	SubjectID   model.SubjectID              `json:"subject_id"`
	VolStatus   uint32                       `json:"vol_status"`
	EpStatus    uint32                       `json:"ep_status"`
	SubjectType uint8                        `json:"subject_type"`
	Type        collection.SubjectCollection `json:"type"`
	Rate        uint8                        `json:"rate"`
	Private     bool                         `json:"private"`
}

func ConvertModelSubjectCollection(c collection.UserSubjectCollection, subject SlimSubjectV0) SubjectCollection {
	return SubjectCollection{
		SubjectID:   c.SubjectID,
		SubjectType: c.SubjectType,
		Rate:        c.Rate,
		Type:        c.Type,
		Tags:        c.Tags,
		EpStatus:    c.EpStatus,
		VolStatus:   c.VolStatus,
		UpdatedAt:   c.UpdatedAt,
		Private:     c.Private,
		Comment:     null.NilString(c.Comment),
		Subject:     subject,
	}
}

type PersonCollection struct {
	ID        uint32       `json:"id"`
	Type      uint8        `json:"type"`
	Name      string       `json:"name"`
	Images    PersonImages `json:"images"`
	CreatedAt time.Time    `json:"created_at"`
}

func ConvertModelPersonCollection(c collection.UserPersonCollection, person model.Person) PersonCollection {
	img := PersonImage(person.Image)
	return PersonCollection{
		ID:        person.ID,
		Type:      person.Type,
		Name:      person.Name,
		Images:    img,
		CreatedAt: c.CreatedAt,
	}
}

func ConvertModelCharacterCollection(c collection.UserPersonCollection, character model.Character) PersonCollection {
	img := PersonImage(character.Image)
	return PersonCollection{
		ID:        character.ID,
		Type:      character.Type,
		Name:      character.Name,
		Images:    img,
		CreatedAt: c.CreatedAt,
	}
}
