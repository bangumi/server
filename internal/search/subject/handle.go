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
package subject

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	wiki "github.com/bangumi/wiki-parser-go"
	"github.com/labstack/echo/v4"
	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/compat"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/tag"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

const defaultLimit = 10
const maxLimit = 20

type Req struct {
	Keyword string    `json:"keyword"`
	Sort    string    `json:"sort"`
	Filter  ReqFilter `json:"filter"`
}

type ReqFilter struct { //nolint:musttag
	Type     []model.SubjectType `json:"type"`      // or
	Tag      []string            `json:"tag"`       // and
	AirDate  []string            `json:"air_date"`  // and
	Score    []string            `json:"rating"`    // and
	Rank     []string            `json:"rank"`      // and
	MetaTags []string            `json:"meta_tags"` // and

	// if NSFW subject is enabled
	NSFW null.Bool `json:"nsfw"`
}

type hit struct {
	ID model.SubjectID `json:"id"`
}

type ReponseSubject struct {
	Date       *string                   `json:"date"`
	Platform   *string                   `json:"platform"`
	Images     res.SubjectImages         `json:"images"`
	Image      string                    `json:"image"`
	Summary    string                    `json:"summary"`
	Name       string                    `json:"name"`
	NameCN     string                    `json:"name_cn"`
	Tags       []res.SubjectTag          `json:"tags"`
	Infobox    res.V0wiki                `json:"infobox"`
	Rating     res.Rating                `json:"rating"`
	Collection res.SubjectCollectionStat `json:"collection"`
	ID         model.SubjectID           `json:"id"`
	Eps        uint32                    `json:"eps"`
	MetaTags   []string                  `json:"meta_tags"`
	Volumes    uint32                    `json:"volumes"`
	Series     bool                      `json:"series"`
	Locked     bool                      `json:"locked"`
	NSFW       bool                      `json:"nsfw"`
	TypeID     model.SubjectType         `json:"type"`
	Redirect   model.SubjectID           `json:"-"`
}

//nolint:funlen
func (c *client) Handle(ctx echo.Context) error {
	auth := accessor.GetFromCtx(ctx)
	q, err := req.GetPageQuerySoftLimit(ctx, defaultLimit, maxLimit)
	if err != nil {
		return err
	}

	var r Req
	if err = json.NewDecoder(ctx.Request().Body).Decode(&r); err != nil {
		return res.JSONError(ctx, err)
	}

	if !auth.AllowNSFW() {
		r.Filter.NSFW = null.Bool{Set: true, Value: false}
	}

	result, err := c.doSearch(r.Keyword, filterToMeiliFilter(r.Filter), r.Sort, q.Limit, q.Offset)
	if err != nil {
		return errgo.Wrap(err, "search")
	}

	var hits []hit
	if err = json.Unmarshal(result.Hits, &hits); err != nil {
		return errgo.Wrap(err, "json.Unmarshal")
	}
	ids := slice.Map(hits, func(h hit) model.SubjectID { return h.ID })

	subjects, err := c.repo.GetByIDs(ctx.Request().Context(), ids, subject.Filter{})
	if err != nil {
		return errgo.Wrap(err, "subjectRepo.GetByIDs")
	}

	var data = make([]ReponseSubject, 0, len(subjects))
	for _, id := range ids {
		s, ok := subjects[id]
		if !ok {
			continue
		}
		var metaTags []tag.Tag

		for _, t := range strings.Split(s.MetaTags, " ") {
			if t == "" {
				continue
			}
			metaTags = append(metaTags, tag.Tag{Name: t, Count: 1})
		}

		subject := toResponseSubject(s, metaTags)
		data = append(data, subject)
	}

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
		sortOpt = []string{"rank:asc"}
	default:
		return nil, res.BadRequest("sort not supported")
	}

	raw, err := c.index.SearchRaw(words, &meilisearch.SearchRequest{
		Offset: int64(offset),
		Limit:  int64(limit),
		Filter: filter,
		Sort:   sortOpt,
	})
	if err != nil {
		return nil, errgo.Wrap(err, "meilisearch search")
	}

	var r meiliSearchResponse
	if err := json.Unmarshal(*raw, &r); err != nil {
		return nil, errgo.Wrap(err, "json.Unmarshal")
	}

	return &r, nil
}

type meiliSearchResponse struct {
	Hits               json.RawMessage `json:"hits"`
	EstimatedTotalHits int64           `json:"estimatedTotalHits"` //nolint:tagliatelle
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
		filter = append(filter, []string{fmt.Sprintf("nsfw = %t", req.NSFW.Value)})
	}

	for _, tag := range req.MetaTags {
		filter = append(filter, []string{"meta_tag = " + strconv.Quote(tag)})
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
	if !isDigitsOnly(date[:4]) ||
		date[4] != '-' || !isDigitsOnly(date[5:7]) ||
		date[7] != '-' || !isDigitsOnly(date[8:10]) {
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

func toResponseSubject(s model.Subject, metaTags []tag.Tag) ReponseSubject {
	images := res.SubjectImage(s.Image)
	return ReponseSubject{
		ID:       s.ID,
		Image:    images.Large,
		Images:   images,
		Summary:  s.Summary,
		Name:     s.Name,
		Platform: res.PlatformString(s),
		NameCN:   s.NameCN,
		Date:     null.NilString(s.Date),
		Infobox:  compat.V0Wiki(wiki.ParseOmitError(s.Infobox).NonZero()),
		Volumes:  s.Volumes,
		Redirect: s.Redirect,
		Eps:      s.Eps,
		MetaTags: lo.Map(metaTags, func(item tag.Tag, index int) string {
			return item.Name
		}),
		Tags: slice.Map(s.Tags, func(tag model.Tag) res.SubjectTag {
			return res.SubjectTag{
				Name:  tag.Name,
				Count: tag.Count,
			}
		}),
		Collection: res.SubjectCollectionStat{
			OnHold:  s.OnHold,
			Wish:    s.Wish,
			Dropped: s.Dropped,
			Collect: s.Collect,
			Doing:   s.Doing,
		},
		TypeID: s.TypeID,
		Series: s.Series,
		Locked: s.Locked(),
		NSFW:   s.NSFW,
		Rating: res.Rating{
			Rank:  s.Rating.Rank,
			Total: s.Rating.Total,
			Count: res.Count{
				Field1:  s.Rating.Count.Field1,
				Field2:  s.Rating.Count.Field2,
				Field3:  s.Rating.Count.Field3,
				Field4:  s.Rating.Count.Field4,
				Field5:  s.Rating.Count.Field5,
				Field6:  s.Rating.Count.Field6,
				Field7:  s.Rating.Count.Field7,
				Field8:  s.Rating.Count.Field8,
				Field9:  s.Rating.Count.Field9,
				Field10: s.Rating.Count.Field10,
			},
			Score: s.Rating.Score,
		},
	}
}
