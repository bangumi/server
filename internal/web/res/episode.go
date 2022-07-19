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
	"github.com/bangumi/server/internal/model"
)

type Episode struct {
	Airdate     string          `json:"airdate"`
	Name        string          `json:"name"`
	NameCN      string          `json:"name_cn"`
	Duration    string          `json:"duration"`
	Description string          `json:"desc"`
	Ep          float32         `json:"ep"`
	Sort        float32         `json:"sort"`
	ID          model.EpisodeID `json:"id"`
	SubjectID   model.SubjectID `json:"subject_id"`
	Comment     uint32          `json:"comment"`
	Type        model.EpType    `json:"type"`
	Disc        uint8           `json:"disc"`
}

func ConvertModelEpisode(s model.Episode) Episode {
	return Episode{
		ID:          s.ID,
		Name:        s.Name,
		NameCN:      s.NameCN,
		Ep:          s.Ep,
		Sort:        s.Sort,
		Duration:    s.Duration,
		Airdate:     s.Airdate,
		SubjectID:   s.SubjectID,
		Description: s.Description,
		Comment:     s.Comment,
		Type:        s.Type,
		Disc:        s.Disc,
	}
}
