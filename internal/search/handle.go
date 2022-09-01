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

// Package search 基于 meilisearch 提供搜索功能
package search

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/meilisearch/meilisearch-go"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/obj"
	"github.com/bangumi/server/internal/web/accessor"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

type Client interface {
	Handler
	OnSubjectUpdate(ctx context.Context, id model.SubjectID) error
	OnSubjectDelete(ctx context.Context, id model.SubjectID) error
}

// Handler
// TODO: 想个办法挪到 web 里面去.
type Handler interface {
	Handle(ctx *fiber.Ctx, auth *accessor.Accessor) error
}

const defaultLimit = 50
const maxLimit = 200

type Req struct {
	Keyword string    `json:"keyword"`
	Sort    string    `json:"sort"`
	Filter  ReqFilter `json:"filter"`
}

type ReqFilter struct {
	Type    []model.SubjectType `json:"type"`     // or
	Tag     []string            `json:"tag"`      // and
	AirDate []string            `json:"air_date"` // and
	Score   []string            `json:"rating"`   // and
	Rank    []string            `json:"rank"`     // and
	NSFW    null.Bool           `json:"nsfw"`
}

func (c *client) Handle(ctx *fiber.Ctx, auth *accessor.Accessor) error {
	q, err := req.GetPageQuery(ctx, defaultLimit, maxLimit)
	if err != nil {
		return err
	}

	var r Req
	if err = json.Unmarshal(ctx.Body(), &r); err != nil {
		return res.JSONError(ctx, err)
	}

	if !auth.AllowNSFW() {
		r.Filter.NSFW = null.New(false)
	}

	result, err := c.doSearch(r.Keyword, filterToMeiliFilter(r.Filter), r.Sort, q.Limit, q.Offset)
	if err != nil {
		return errgo.Wrap(err, "search")
	}

	data := make([]resSubject, 0, len(result.Hits))
	for _, h := range result.Hits {
		hit, ok := h.(map[string]any)
		if !ok {
			continue
		}

		recordMap, ok := hit["record"].(map[string]any)
		if !ok {
			continue
		}

		data = append(data, resSubject{
			ID:      model.SubjectID(obj.GetFloat64(recordMap, "id")),
			Summary: obj.GetString(hit, "summary"),
			Date:    obj.GetString(recordMap, "date"),
			Image:   res.SubjectImage(obj.GetString(recordMap, "image")).Large,
			Name:    obj.GetString(recordMap, "name"),
			NameCN:  obj.GetString(recordMap, "name_cn"),
			Tags:    anyToResTag(recordMap["tags"]),
			Score:   obj.GetFloat64(recordMap, "score"),
			Rank:    uint32(obj.GetFloat64(recordMap, "rank")),
		})
	}

	return res.JSON(ctx, res.Paged{
		Data:   data,
		Total:  result.EstimatedTotalHits,
		Limit:  q.Limit,
		Offset: q.Offset,
	})
}

func anyToResTag(v any) []res.SubjectTag {
	if v == nil {
		return nil
	}

	tags, ok := v.([]any)
	if !ok {
		return nil
	}

	var result = make([]res.SubjectTag, 0, len(tags))
	for _, tagAny := range tags {
		tagMap, ok := tagAny.(map[string]any)
		if !ok {
			continue
		}
		tag := res.SubjectTag{
			Name:  obj.GetString(tagMap, "name"),
			Count: int(obj.GetFloat64(tagMap, "count")),
		}

		result = append(result, tag)
	}

	return result
}

func (c *client) doSearch(
	words string,
	filter [][]string,
	sort string,
	limit, offset int,
) (*meilisearch.SearchResponse, error) {
	if limit == 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	var sortOpt []string
	switch sort {
	case "", "match":
	case "score":
		sortOpt = []string{"score:desc"}
	case "heat":
		sortOpt = []string{"heat:desc"}
	case "rank":
		sortOpt = []string{"rank:desc"}
	default:
		return nil, res.BadRequest("sort not supported")
	}

	response, err := c.subjectIndex.Search(words, &meilisearch.SearchRequest{
		Offset: int64(offset),
		Limit:  int64(limit),
		Filter: filter,
		Sort:   sortOpt,
	})
	if err != nil {
		return nil, errgo.Wrap(err, "meilisearch search")
	}

	return response, nil
}

type resSubject struct {
	Date    string           `json:"date"`
	Image   string           `json:"image"`
	Name    string           `json:"name"`
	NameCN  string           `json:"name_cn"`
	Tags    []res.SubjectTag `json:"tags,omitempty"`
	Score   float64          `json:"score"`
	ID      model.SubjectID  `json:"id"`
	Rank    uint32           `json:"rank"`
	Summary string           `json:"summary"`
}

func filterToMeiliFilter(req ReqFilter) [][]string {
	var filter = make([][]string, 0, 5+len(req.Tag))

	if len(req.AirDate) != 0 {
		filter = append(filter, parseDateFilter(req.AirDate))
	}

	for _, tag := range req.Tag {
		filter = append(filter, []string{"tag = " + strconv.Quote(tag)})
	}

	if len(req.Type) != 0 {
		filter = append(filter, slice.Map(req.Type, func(s model.SubjectType) string {
			return fmt.Sprintf("type = %d", s)
		}))
	}

	if len(req.Rank) != 0 {
		filter = append(filter, slice.Map(req.Rank, func(s string) string {
			return "rank" + s
		}))
	}

	if len(req.Score) != 0 {
		filter = append(filter, slice.Map(req.Score, func(s string) string {
			return "score " + s
		}))
	}

	if req.NSFW.Set {
		filter = append(filter, []string{"nsfw = " + strconv.FormatBool(req.NSFW.Value)})
	}

	return filter
}

// parse date filter like `<2020-01-20`, `>=2020-01-23`.
func parseDateFilter(filters []string) []string {
	var result = make([]string, 0, len(filters))

	for _, s := range filters {
		switch {
		case strings.HasPrefix(s, ">="):
			if v, ok := parseDateValOk(s[2:]); ok {
				result = append(result, fmt.Sprintf("date >= %d", v))
			}
		case strings.HasPrefix(s, ">"):
			if v, ok := parseDateValOk(s[1:]); ok {
				result = append(result, fmt.Sprintf("date > %d", v))
			}
		case strings.HasPrefix(s, "<="):
			if v, ok := parseDateValOk(s[2:]); ok {
				result = append(result, fmt.Sprintf("date <= %d", v))
			}
		case strings.HasPrefix(s, "<"):
			if v, ok := parseDateValOk(s[1:]); ok {
				result = append(result, fmt.Sprintf("date < %d", v))
			}
		default:
			if v, ok := parseDateValOk(s); ok {
				result = append(result, fmt.Sprintf("date = %d", v))
			}
		}
	}

	return result
}

func parseDateValOk(date string) (int, bool) {
	if len(date) < 10 {
		return 0, false
	}

	// 2008-10-05 format
	if !(isDigitsOnly(date[:4]) &&
		date[4] == '-' &&
		isDigitsOnly(date[5:7]) &&
		date[7] == '-' &&
		isDigitsOnly(date[8:10])) {
		return 0, false
	}

	v, err := strconv.Atoi(date[:4])
	if err != nil {
		return 0, false
	}
	val := v * 10000

	v, err = strconv.Atoi(date[5:7])
	if err != nil {
		return 0, false
	}
	val += v * 100

	v, err = strconv.Atoi(date[8:10])
	if err != nil {
		return 0, false
	}
	val += v

	return val, true
}

func isDigitsOnly(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
