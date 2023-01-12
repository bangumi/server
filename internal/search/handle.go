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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"github.com/labstack/echo/v4"
	"github.com/meilisearch/meilisearch-go"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

type Client interface {
	Handler
	OnSubjectUpdate(ctx context.Context, id model.SubjectID) error
	Close()
	OnSubjectDelete(ctx context.Context, id model.SubjectID) error
}

// Handler
// TODO: 想个办法挪到 web 里面去.
type Handler interface {
	Handle(c echo.Context) error
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

func (c *client) Handle(ctx echo.Context) error {
	auth := accessor.GetFromCtx(ctx)
	q, err := req.GetPageQuery(ctx, defaultLimit, maxLimit)
	if err != nil {
		return err
	}

	var r Req
	if err = decoder.NewStreamDecoder(ctx.Request().Body).Decode(&r); err != nil {
		return res.JSONError(ctx, err)
	}

	if !auth.AllowNSFW() {
		r.Filter.NSFW = null.New(false)
	}

	result, err := c.doSearch(r.Keyword, filterToMeiliFilter(r.Filter), r.Sort, q.Limit, q.Offset)
	if err != nil {
		return errgo.Wrap(err, "search")
	}

	ids := make([]model.SubjectID, 0, len(result.Hits))
	for _, h := range result.Hits {
		var hit struct {
			ID model.SubjectID `json:"id"`
		}

		if err = sonic.Unmarshal(h, &hit); err != nil {
			return errgo.Wrap(err, "json.Unmarshal")
		}

		ids = append(ids, hit.ID)
	}

	subjects, err := c.subjectRepo.GetByIDs(ctx.Request().Context(), ids, subject.Filter{NSFW: r.Filter.NSFW})
	if err != nil {
		return errgo.Wrap(err, "subjectRepo.GetByIDs")
	}

	data := slice.Map(ids, func(id model.SubjectID) Record {
		s := subjects[id]

		return Record{
			Date:   s.Date,
			Image:  res.SubjectImage(s.Image).Large,
			Name:   s.Name,
			NameCN: s.NameCN,
			Tags: slice.Map(s.Tags, func(item model.Tag) res.SubjectTag {
				return res.SubjectTag{Name: item.Name, Count: item.Count}
			}),
			Score: s.Rating.Score,
			ID:    s.ID,
			Rank:  s.Rating.Rank,
		}
	})

	return ctx.JSON(http.StatusOK, res.Paged{
		Data:   data,
		Total:  result.EstimatedTotalHits,
		Limit:  q.Limit,
		Offset: q.Offset,
	})
}

func (c *client) doSearch(
	words string,
	filter [][]string,
	sort string,
	limit, offset int,
) (*meiliSearchResponse, error) {
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

	raw, err := c.subjectIndex.SearchRaw(words, &meilisearch.SearchRequest{
		Offset: int64(offset),
		Limit:  int64(limit),
		Filter: filter,
		Sort:   sortOpt,
	})
	if err != nil {
		return nil, errgo.Wrap(err, "meilisearch search")
	}

	var r meiliSearchResponse
	if err := sonic.Unmarshal(*raw, &r); err != nil {
		return nil, errgo.Wrap(err, "json.Unmarshal")
	}

	return &r, nil
}

type meiliSearchResponse struct {
	Hits               []json.RawMessage `json:"hits"`
	EstimatedTotalHits int64             `json:"estimatedTotalHits"` //nolint:tagliatelle
}

func filterToMeiliFilter(req ReqFilter) [][]string {
	var filter = make([][]string, 0, 5+len(req.Tag))

	// OR

	if len(req.AirDate) != 0 {
		filter = append(filter, parseDateFilter(req.AirDate)...)
	}

	if len(req.Type) != 0 {
		filter = append(filter, slice.Map(req.Type, func(s model.SubjectType) string {
			return fmt.Sprintf("type = %d", s)
		}))
	}
	if req.NSFW.Set {
		filter = append(filter, []string{"nsfw = " + strconv.FormatBool(req.NSFW.Value)})
	}

	// AND

	for _, tag := range req.Tag {
		filter = append(filter, []string{"tag = " + strconv.Quote(tag)})
	}

	for _, s := range req.Rank {
		filter = append(filter, []string{"rank" + s})
	}

	for _, s := range req.Score {
		filter = append(filter, []string{"score " + s})
	}

	return filter
}

// parse date filter like `<2020-01-20`, `>=2020-01-23`.
func parseDateFilter(filters []string) [][]string {
	var result = make([][]string, 0, len(filters))

	for _, s := range filters {
		switch {
		case strings.HasPrefix(s, ">="):
			if v, ok := parseDateValOk(s[2:]); ok {
				result = append(result, []string{fmt.Sprintf("date >= %d", v)})
			}
		case strings.HasPrefix(s, ">"):
			if v, ok := parseDateValOk(s[1:]); ok {
				result = append(result, []string{fmt.Sprintf("date > %d", v)})
			}
		case strings.HasPrefix(s, "<="):
			if v, ok := parseDateValOk(s[2:]); ok {
				result = append(result, []string{fmt.Sprintf("date <= %d", v)})
			}
		case strings.HasPrefix(s, "<"):
			if v, ok := parseDateValOk(s[1:]); ok {
				result = append(result, []string{fmt.Sprintf("date < %d", v)})
			}
		default:
			if v, ok := parseDateValOk(s); ok {
				result = append(result, []string{fmt.Sprintf("date = %d", v)})
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
