package search

import (
	"fmt"
	"strconv"
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
		Offset: 0,
		Limit:  0,
		Filter: nil,
		Sort:   nil,
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

	return nil
}

func parseDateValOk(date string) (int, bool) {
	if len(date) < 10 {
		return 0, false
	}

	// 2008-10-05 format
	if !(isDigitsOnly(date[:4]) && date[4] == '-' && isDigitsOnly(date[5:7]) && date[7] == '-' && isDigitsOnly(date[8:10])) {
		return 0, false
	}

	var val = 0

	v, err := strconv.Atoi(date[:4])
	if err != nil {
		return 0, false
	}
	val = v * 10000

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
