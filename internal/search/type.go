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

package search

import (
	"github.com/bangumi/server/internal/model"
)

// Subject 是最终 meilisearch 索引的文档
type Subject struct {
	Summary  string   `json:"summary"`
	Tag      []string `json:"tag,omitempty"`
	Name     []string `json:"name"`
	Record   Record   `json:"record"`
	Date     int      `json:"date,omitempty"`
	Score    float64  `json:"score"`
	PageRank float64  `json:"page_rank,omitempty"`
	Heat     uint32   `json:"heat,omitempty"`
	Rank     uint32   `json:"rank"`
	Platform uint16   `json:"platform,omitempty"`
	Type     uint8    `json:"type"`
	NSFW     bool     `json:"nsfw"`
}

type Record struct {
	Date   string          `json:"date"`
	Image  string          `json:"image"`
	Name   string          `json:"name"`
	NameCN string          `json:"name_cn"`
	Tags   []model.Tag     `json:"tags"`
	Score  float64         `json:"score"`
	ID     model.SubjectID `json:"id"`
	Rank   uint32          `json:"rank"`
}
