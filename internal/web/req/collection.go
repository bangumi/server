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

package req

import (
	"github.com/bangumi/server/internal/pkg/null"
)

type PatchSubjectCollection struct {
	Comment   null.String `json:"comment"`
	Tags      []string    `json:"tags"`
	EpStatus  null.Uint32 `json:"ep_status"`
	VolStatus null.Uint32 `json:"vol_status"`
	Type      null.Uint8  `json:"type" validate:"lte=5,gte=1,omitempty"`
	Private   null.Bool   `json:"private"`
}
