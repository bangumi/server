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

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
)

type SubjectCollection struct {
	UpdatedAt   time.Time                   `json:"updated_at"`
	Comment     *string                     `json:"comment"`
	Tags        []string                    `json:"tags"`
	Subject     SlimSubjectV0               `json:"subject"`
	SubjectID   model.SubjectID             `json:"subject_id"`
	VolStatus   uint32                      `json:"vol_status"`
	EpStatus    uint32                      `json:"ep_status"`
	SubjectType uint8                       `json:"subject_type"`
	Type        model.SubjectCollectionType `json:"type"`
	Rate        uint8                       `json:"rate"`
	Private     bool                        `json:"private"`
}

func ConvertModelSubjectCollection(c model.SubjectCollection, subject SlimSubjectV0) SubjectCollection {
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
