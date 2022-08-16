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
	"strconv"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/pkg/wiki"
)

// 是最终 meilisearch 索引的文档
type subjectIndex struct {
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

func extractSubject(s *model.Subject) subjectIndex {
	tags := s.Tags

	w := wiki.ParseOmitError(s.Infobox)

	rank := s.Rating.Total
	score := s.Rating.Score

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	return subjectIndex{
		Name:     extractNames(s, w),
		Tag:      tagNames,
		Summary:  s.Summary,
		NSFW:     s.NSFW,
		Type:     s.TypeID,
		Date:     parseDateVal(s.Date),
		Platform: s.PlatformID,
		PageRank: float64(rank),
		Rank:     s.Rating.Rank,
		Heat:     heat(s),
		Score:    score,
		Record: Record{
			ID:     s.ID,
			Image:  s.Image,
			Name:   s.Name,
			NameCN: s.NameCN,
			Date:   s.Date,
			Tags:   tags,
			Rank:   rank,
			Score:  score,
		},
	}
}

func parseDateVal(date string) int {
	if len(date) < 10 {
		return 0
	}

	// 2008-10-05 format
	v, err := strconv.Atoi(date[:4])
	if err != nil {
		return 0
	}
	val := v * 10000

	v, err = strconv.Atoi(date[5:7])
	if err != nil {
		return 0
	}
	val += v * 100

	v, err = strconv.Atoi(date[8:10])
	if err != nil {
		return 0
	}
	val += v

	return val
}
