package search

import (
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/meilisearch/meilisearch-go"

	"github.com/bangumi/server/internal/model"
)

type Query struct {
	Q     string   `query:"q"`
	Sort  string   `query:"sort"`
	Since []string `query:"since"`
	Limit int64    `query:"limit"`
}

func (c *Client) Handler() fiber.Handler {
	search := c.search

	return func(c *fiber.Ctx) error {
		query := Query{}
		if err := c.QueryParser(&query); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid query string")
		}

		query.Q = strings.TrimSpace(query.Q)
		if query.Q == "" {
			return c.SendString("empty query string")
		}

		w, q, err := parse(query.Q)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(err)
		}

		runtime.KeepAlive(w)
		runtime.KeepAlive(q)
		runtime.KeepAlive(search)

		// result, err := buildService(search, query, q)
		// if err != nil {
		// 	return errgo.Wrap(err, "search")
		// }
		//
		// res := Response{
		// 	Pagination: Pagination{Limit: query.Limit},
		// 	Data:       make([]resSubject, len(result.Hits.Hits)),
		// }
		//
		// if len(result.Hits.Hits) > 0 {
		// 	res.Pagination.Since = result.Hits.Hits[len(result.Hits.Hits)-1].Sort
		// }
		//
		// for i, hit := range result.Hits.Hits {
		// 	var source = Subject{}
		// 	err = json.Unmarshal(hit.Source, &source)
		// 	if err != nil {
		// 		return err
		// 	}
		//
		// 	res.Data[i] = resSubject{
		// 		ID:     source.Record.ID,
		// 		Date:   source.Date,
		// 		Image:  source.Record.Image,
		// 		Name:   source.Record.Name,
		// 		NameCN: source.Record.NameCN,
		// 		Tags:   source.Record.Tags,
		// 		Score:  source.Record.Score,
		// 		Rank:   source.Record.Rank,
		// 	}
		// }
		//
		// return c.JSON(res)
		return nil
	}
}

func buildService(es *meilisearch.Client, query Query) *meilisearch.SearchResponse {
	if query.Limit == 0 {
		query.Limit = 10
	} else if query.Limit > 50 {
		query.Limit = 50
	}

	service, err := es.Index("subjects").Search(query.Q, &meilisearch.SearchRequest{
		Limit: query.Limit,
	})
	if err != nil {
		panic(err)
	}

	// switch query.Sort {
	// case "":
	// 	service = service.SortBy(elastic.NewFieldSort("_score").Desc(), elastic.NewFieldSort("_id"))
	// case "rank":
	// 	service = service.SortBy(elastic.NewFieldSort("rank").Desc(), elastic.NewFieldSort("_id"))
	// case "-rank":
	// 	service = service.SortBy(elastic.NewFieldSort("rank"), elastic.NewFieldSort("_id"))
	// case "airdate":
	// 	service = service.SortBy(elastic.NewFieldSort("date"), elastic.NewFieldSort("_id"))
	// case "-airdate":
	// 	service = service.SortBy(elastic.NewFieldSort("date").Desc(), elastic.NewFieldSort("_id"))
	// }
	//
	// if len(query.Since) > 0 {
	// 	service = service.SearchAfter(toInterface(query.Since)...)
	// }

	// return service.Query(q)
	return service
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
