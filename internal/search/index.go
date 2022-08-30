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
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/pkg/wiki"
)

// 最终 meilisearch 索引的文档.
// 使用 `searchable:"true"`，`filterable:"true"`， `sortable:"true"`
// 三种 tag 来设置是否可以被搜索，索引和排序.
type subjectIndex struct {
	ID       model.SubjectID `json:"id"`
	Summary  string          `json:"summary" searchable:"true"`
	Tag      []string        `json:"tag,omitempty" searchable:"true"`
	Name     []string        `json:"name" searchable:"true"`
	Record   Record          `json:"record"`
	Date     int             `json:"date,omitempty" filterable:"true" sortable:"true"`
	Score    float64         `json:"score" filterable:"true" sortable:"true"`
	PageRank float64         `json:"page_rank" sortable:"true"`
	Heat     uint32          `json:"heat" sortable:"true"`
	Rank     uint32          `json:"rank" filterable:"true" sortable:"true"`
	Platform uint16          `json:"platform,omitempty"`
	Type     uint8           `json:"type" filterable:"true"`
	NSFW     bool            `json:"nsfw" filterable:"true"`
}

type Record struct {
	Date   string           `json:"date"`
	Image  string           `json:"image"`
	Name   string           `json:"name"`
	NameCN string           `json:"name_cn"`
	Tags   []res.SubjectTag `json:"tags"`
	Score  float64          `json:"score"`
	ID     model.SubjectID  `json:"id"`
	Rank   uint32           `json:"rank"`
}

func extractSubject(s *model.Subject) subjectIndex {
	tags := s.Tags

	w := wiki.ParseOmitError(s.Infobox)

	score := s.Rating.Score

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	return subjectIndex{
		ID:       s.ID,
		Name:     extractNames(s, w),
		Tag:      tagNames,
		Summary:  s.Summary,
		NSFW:     s.NSFW,
		Type:     s.TypeID,
		Date:     parseDateVal(s.Date),
		Platform: s.PlatformID,
		PageRank: float64(s.Rating.Total),
		Rank:     s.Rating.Rank,
		Heat:     heat(s),
		Score:    score,
		Record: Record{
			ID:     s.ID,
			Image:  s.Image,
			Name:   s.Name,
			NameCN: s.NameCN,
			Date:   s.Date,
			Tags: slice.Map(tags, func(t model.Tag) res.SubjectTag {
				return res.SubjectTag{Name: t.Name, Count: t.Count}
			}),
			Rank:  s.Rating.Rank,
			Score: score,
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
