package search

import (
	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"

	"github.com/bangumi/server/internal/model"
)

type Query struct {
	Q     string   `query:"q"`
	Sort  string   `query:"sort"`
	Since []string `query:"since"`
	Limit int      `query:"limit"`
}

func (c *Client) Handler() fiber.Handler {
	es := c.es

	return func(c *fiber.Ctx) error {
		query := Query{}
		if err := c.QueryParser(&query); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid query string")
		}

		query.Q = strings.TrimSpace(query.Q)
		if query.Q == "" {
			return c.SendString("empty query string")
		}

		q, err := parseQueryLine(query.Q)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(err)
		}

		result, err := buildService(es, query, q).Do(c.Context())
		if err != nil {
			return errors.Wrap(err, "es")
		}

		res := Response{
			Pagination: Pagination{Limit: query.Limit},
			Data:       make([]resSubject, len(result.Hits.Hits)),
		}

		if len(result.Hits.Hits) > 0 {
			res.Pagination.Since = result.Hits.Hits[len(result.Hits.Hits)-1].Sort
		}

		for i, hit := range result.Hits.Hits {
			var source = Subject{}
			err = json.Unmarshal(hit.Source, &source)
			if err != nil {
				return err
			}

			res.Data[i] = resSubject{
				ID:     source.Record.ID,
				Date:   source.Date,
				Image:  source.Record.Image,
				Name:   source.Record.Name,
				NameCN: source.Record.NameCN,
				Tags:   source.Record.Tags,
				Score:  source.Record.Score,
				Rank:   source.Record.Rank,
			}
		}

		return c.JSON(res)
	}
}

func buildService(es *elastic.Client, query Query, q *elastic.BoolQuery) *elastic.SearchService {
	if query.Limit == 0 {
		query.Limit = 10
	} else if query.Limit > 50 {
		query.Limit = 50
	}

	service := es.Search("subjects").Size(query.Limit)

	switch query.Sort {
	case "":
		service = service.SortBy(elastic.NewFieldSort("_score").Desc(), elastic.NewFieldSort("_id"))
	case "rank":
		service = service.SortBy(elastic.NewFieldSort("rank").Desc(), elastic.NewFieldSort("_id"))
	case "-rank":
		service = service.SortBy(elastic.NewFieldSort("rank"), elastic.NewFieldSort("_id"))
	case "airdate":
		service = service.SortBy(elastic.NewFieldSort("date"), elastic.NewFieldSort("_id"))
	case "-airdate":
		service = service.SortBy(elastic.NewFieldSort("date").Desc(), elastic.NewFieldSort("_id"))
	}

	if len(query.Since) > 0 {
		service = service.SearchAfter(toInterface(query.Since)...)
	}

	return service.Query(q)
}

func parseQueryLine(s string) (*elastic.BoolQuery, error) {
	b, err := parse(s)
	if err != nil {
		return nil, err
	}

	b.Must(
		elastic.NewRankFeatureQuery("heat").Boost(5),        //nolint:gomnd
		elastic.NewRankFeatureQuery("page_rank").Boost(2.5), //nolint:gomnd
	).MinimumShouldMatch("75%")

	return b, nil
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
