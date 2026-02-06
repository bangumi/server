package character

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

const defaultLimit = 10
const maxLimit = 20

type Req struct {
	Keyword string    `json:"keyword"`
	Filter  ReqFilter `json:"filter"`
}

type ReqFilter struct { //nolint:musttag
	NSFW null.Bool `json:"nsfw"`
}

type hit struct {
	ID model.CharacterID `json:"id"`
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

	result, err := c.doSearch(r.Keyword, filterToMeiliFilter(r.Filter), q.Limit, q.Offset)
	if err != nil {
		return errgo.Wrap(err, "search")
	}

	var hits []hit
	if err = json.Unmarshal(result.Hits, &hits); err != nil {
		return errgo.Wrap(err, "json.Unmarshal")
	}
	ids := slice.Map(hits, func(h hit) model.SubjectID { return h.ID })

	characters, err := c.repo.GetByIDs(ctx.Request().Context(), ids)
	if err != nil {
		return errgo.Wrap(err, "characterRepo.GetByIDs")
	}

	var data = make([]res.CharacterV0, 0, len(characters))
	for _, id := range ids {
		s, ok := characters[id]
		if !ok {
			continue
		}
		character := res.ConvertModelCharacter(s)
		data = append(data, character)
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
	limit, offset int,
) (*meiliSearchResponse, error) {
	if limit == 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	raw, err := c.index.SearchRaw(words, &meilisearch.SearchRequest{
		Offset: int64(offset),
		Limit:  int64(limit),
		Filter: filter,
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
	var filter = make([][]string, 0, 1)

	if req.NSFW.Set {
		filter = append(filter, []string{fmt.Sprintf("nsfw = %t", req.NSFW.Value)})
	}

	return filter
}
