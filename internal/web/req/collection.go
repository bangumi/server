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
	"github.com/volatiletech/null/v9"

	"github.com/bangumi/server/internal/model"
)

type NullCollectionType = null.Uint8

type PutEpisodeCollection struct {
	ID   model.EpisodeID    `json:"id"`
	Type NullCollectionType `json:"type"`
}

type PutSubjectCollection struct {
	Comment   string               `json:"comment"`
	Tags      []string             `json:"tags"`
	EpStatus  uint32               `json:"ep_status"`
	VolStatus uint32               `json:"vol_status"`
	Type      model.CollectionType `json:"type"`
	Rate      uint8                `json:"rate"`
	Private   bool                 `json:"private"`
}

type PatchSubjectCollection struct {
	Comment   null.String        `json:"comment"`
	Tags      []string           `json:"tags"`
	EpStatus  uint32             `json:"ep_status"`
	VolStatus uint32             `json:"vol_status"`
	Type      NullCollectionType `json:"type"`
	Rate      null.Uint8         `json:"rate"`
	Private   null.Bool          `json:"private"`
}
