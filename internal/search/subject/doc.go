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

package subject

import (
	"strconv"
	"strings"

	wiki "github.com/bangumi/wiki-parser-go"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/search/searcher"
)

// 最终 meilisearch 索引的文档.
// 使用 `filterable:"true"`， `sortable:"true"`
// 两种 tag 来设置是否可以被索引和排序.
// 搜索字段因为带有排序，所以定义在 [search.searchAbleAttribute] 中.
type document struct {
	ID          model.SubjectID `json:"id"`
	Tag         []string        `json:"tag,omitempty" filterable:"true"`
	MetaTags    []string        `json:"meta_tag" filterable:"true"`
	Name        string          `json:"name" searchable:"true"`
	Aliases     []string        `json:"aliases,omitempty" searchable:"true"`
	Date        int             `json:"date,omitempty" filterable:"true" sortable:"true"`
	Score       float64         `json:"score" filterable:"true" sortable:"true"`
	RatingCount uint32          `json:"rating_count" filterable:"true" sortable:"true"`
	PageRank    float64         `json:"page_rank" sortable:"true"`
	Heat        uint32          `json:"heat" sortable:"true"`
	Rank        uint32          `json:"rank" filterable:"true" sortable:"true"`
	Platform    uint16          `json:"platform,omitempty"`
	Type        uint8           `json:"type" filterable:"true"`
	NSFW        bool            `json:"nsfw" filterable:"true"`
}

func (d *document) GetID() string {
	return strconv.FormatUint(uint64(d.ID), 10)
}

func rankRule() *[]string {
	return &[]string{
		// 相似度最优先
		"exactness",
		"words",
		"typo",
		"proximity",
		"attribute",
		"sort",
		// id 在前的优先展示，主要是为了系列作品能有个很好的顺序
		"id:asc",
		// 以下酌情，我选择优先展示排行榜排名更高、评分更高的条目，且尽量优先展示 sfw 内容
		"rank:asc",
		"score:desc",
		"nsfw:asc",
	}
}

func heat(s *model.Subject) uint32 {
	return s.OnHold + s.Doing + s.Dropped + s.Wish + s.Collect
}

func extract(s *model.Subject) searcher.Document {
	tags := s.Tags

	w := wiki.ParseOmitError(s.Infobox)

	score := s.Rating.Score

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	return &document{
		ID:          s.ID,
		Name:        s.Name,
		Aliases:     extractAliases(s, w),
		MetaTags:    strings.Split(s.MetaTags, " "),
		Tag:         tagNames,
		NSFW:        s.NSFW,
		Type:        s.TypeID,
		Date:        parseDateVal(s.Date),
		Platform:    s.PlatformID,
		RatingCount: s.Rating.Total,
		PageRank:    float64(s.Rating.Total),
		Rank:        s.Rating.Rank,
		Heat:        heat(s),
		Score:       score,
	}
}

func extractAliases(s *model.Subject, w wiki.Wiki) []string {
	var aliases = make([]string, 0, 2)
	if s.NameCN != "" {
		aliases = append(aliases, s.NameCN)
	}

	for _, field := range w.Fields {
		if field.Key == "别名" {
			aliases = append(aliases, searcher.GetWikiValues(field)...)
		}
	}

	return aliases
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
