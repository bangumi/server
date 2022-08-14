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
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/meilisearch/meilisearch-go"
	"github.com/mitchellh/mapstructure"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

const defaultLimit = 50
const maxLimit = 200

type Query struct {
	Q    string
	Sort string
}

func (c *Client) Handle(ctx *fiber.Ctx) error {
	query := Query{
		Q:    strings.TrimSpace(ctx.Query("q")),
		Sort: ctx.Query("sort"),
	}
	if query.Q == "" {
		return ctx.SendString("empty query string")
	}

	q, err := req.GetPageQuery(ctx, defaultLimit, maxLimit)
	if err != nil {
		return err
	}

	w, filter, err := parse(query.Q)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}

	var sort []string
	if query.Sort != "" {
		sort = append(sort, query.Sort)
	}

	result, err := c.doSearch(w, filter, sort, q.Limit, q.Offset)
	if err != nil {
		return errgo.Wrap(err, "search")
	}

	data := make([]resSubject, len(result.Hits))
	for i, hit := range result.Hits {
		var source = Subject{}

		d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			TagName:          "json",
			Result:           &source,
		})
		if err != nil {
			return errgo.Wrap(err, "mapstruct NewDecoder")
		}

		if err := d.Decode(hit); err != nil {
			return errgo.Wrap(err, "failed to convert from any")
		}

		data[i] = resSubject{
			ID:     source.Record.ID,
			Date:   intDateToString(source.Date),
			Image:  source.Record.Image,
			Name:   source.Record.Name,
			NameCN: source.Record.NameCN,
			Tags:   source.Record.Tags,
			Score:  source.Record.Score,
			Rank:   source.Record.Rank,
		}
	}

	return res.JSON(ctx, res.Paged{
		Data:   data,
		Total:  result.EstimatedTotalHits,
		Limit:  q.Limit,
		Offset: q.Offset,
	})
}

func (c *Client) doSearch(
	words string,
	filter [][]string,
	sort []string,
	limit, offset int,
) (*meilisearch.SearchResponse, error) {
	if limit == 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	response, err := c.search.Index("subjects").Search(words, &meilisearch.SearchRequest{
		Offset: int64(offset),
		Limit:  int64(limit),
		Filter: filter,
		Sort:   sort,
	})
	if err != nil {
		return nil, errgo.Wrap(err, "meilisearch search")
	}

	return response, nil
}

type resSubject struct {
	Date   string          `json:"date"`
	Image  string          `json:"image"`
	Name   string          `json:"name"`
	NameCN string          `json:"name_cn"`
	Tags   []model.Tag     `json:"tags,omitempty"`
	Score  float64         `json:"score"`
	ID     model.SubjectID `json:"id"`
	Rank   uint32          `json:"rank"`
}

type Pagination struct {
	Since []interface{} `json:"since"`
	Limit int           `json:"limit"`
}

type Response struct {
	Data       []resSubject `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

func intDateToString(v int) string {
	if v == 0 {
		return ""
	}

	return fmt.Sprintf("%04d-%02d-%02d", v/10000, (v%10000)/100*100, v%100)
}
