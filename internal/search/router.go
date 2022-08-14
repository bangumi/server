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
)

type Query struct {
	Q      string `query:"q"`
	Sort   string `query:"sort"`
	Offset int64  `query:"offset"`
	Limit  int64  `query:"limit"`
}

func (c *Client) Handle(ctx *fiber.Ctx) error {
	query := Query{}
	if err := ctx.QueryParser(&query); err != nil {
		return ctx.Status(fiber.StatusBadRequest).SendString("invalid query string")
	}

	query.Q = strings.TrimSpace(query.Q)
	if query.Q == "" {
		return ctx.SendString("empty query string")
	}

	w, filter, err := parse(query.Q)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}

	var sort []string
	if query.Sort != "" {
		sort = append(sort, query.Sort)
	}

	result, err := c.doSearch(w, filter, sort, query.Limit, query.Offset)
	if err != nil {
		return errgo.Wrap(err, "search")
	}

	res := Response{
		Pagination: Pagination{Limit: query.Limit},
		Data:       make([]resSubject, len(result.Hits)),
	}

	for i, hit := range result.Hits {
		var source = Subject{}
		if err := mapstructure.Decode(hit, &source); err != nil {
			return err
		}

		res.Data[i] = resSubject{
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

	return ctx.JSON(res)
}

func (c *Client) doSearch(
	words string,
	filter [][]string,
	sort []string,
	limit, offset int64,
) (*meilisearch.SearchResponse, error) {
	if limit == 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	response, err := c.search.Index("subjects").Search(words, &meilisearch.SearchRequest{
		Offset: offset,
		Limit:  limit,
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
	Limit int64         `json:"limit"`
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
