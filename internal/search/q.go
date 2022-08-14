package search

import (
	"strings"

	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/search/syntax"
)

func parse(s string) (string, *meilisearch.SearchRequest, error) {
	r, err := syntax.Parse(s)
	if err != nil {
		return "", nil, errgo.Wrap(err, "parse syntax")
	}

	logger.Debug("parse search query", zap.String("input", s), zap.Any("result", r))

	req := &meilisearch.SearchRequest{
		Offset:                0,
		Limit:                 0,
		AttributesToRetrieve:  nil,
		AttributesToCrop:      nil,
		CropLength:            0,
		CropMarker:            "",
		AttributesToHighlight: nil,
		HighlightPreTag:       "",
		HighlightPostTag:      "",
		Filter:                nil,
		ShowMatchesPosition:   false,
		Facets:                nil,
		PlaceholderSearch:     false,
		Sort:                  nil,
	}

	var filter [][]string

	for field, values := range r.Filter {
		var op string
		if field[0] == '-' {
			op = " -= "
		} else {
			op = " = "
		}

		switch field {
		case "airdate":
			// parseDateFilter("date", values, qq)
		case "tag":
			for _, value := range values {
				filter = append(filter, []string{"tag" + op + value})
			}
		case "game_platform":
			filter = append(filter, values)
		case "type":
			filter = append(filter, values)
		}
	}

	return strings.Join(r.Keyword, " "), req, nil
}

// parse date filter like `<2020-01-20`, `>=2020-01-23`.
func parseDateFilter(field string, filters []string) []string {
	// for _, s := range filters {
	// 	switch {
	// 	case strings.HasPrefix(s, ">="):
	// 		((field).Gte(s[2:]))
	// 	case strings.HasPrefix(s, ">"):
	// 		((field).Gt(s[1:]))
	// 	case strings.HasPrefix(s, "<="):
	// 		((field).Lte(s[2:]))
	// 	case strings.HasPrefix(s, "<"):
	// 		((field).Lt(s[1:]))
	// 	default:
	// 		(field, filters)
	// 	}
	// }

	return nil
}
