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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
)

type SubjectCollection struct {
	UpdatedAt   time.Time             `json:"updated_at"`
	Comment     *string               `json:"comment"`
	Tags        []string              `json:"tags"`
	SubjectID   model.SubjectID       `json:"subject_id"`
	EpStatus    uint32                `json:"ep_status"`
	VolStatus   uint32                `json:"vol_status"`
	SubjectType uint8                 `json:"subject_type"`
	Type        domain.CollectionType `json:"type"`
	Rate        uint8                 `json:"rate"`
	Private     bool                  `json:"private"`
}
