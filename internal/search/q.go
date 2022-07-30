package search

import (
	"strings"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/search/syntax"
)

func parse(s string) (*elastic.BoolQuery, error) {
	r, err := syntax.Parse(s)
	if err != nil {
		return nil, errors.Wrap(err, "parse syntax")
	}

	logger.Debug("parse search query", zap.String("input", s), zap.Any("result", r))

	q := elastic.NewBoolQuery()
	for _, w := range r.Keyword {
		if w[0] == '-' {
			q.MustNot(elastic.NewMatchQuery("name", w[1:]))
		} else {
			q.Must(elastic.NewMatchQuery("name", w))
		}
	}

	for field, values := range r.Filter {
		onField(field, values, q)
	}

	return q, nil
}

func onField(field string, values []string, b *elastic.BoolQuery) {
	var n string
	var qq func(filters ...elastic.Query) *elastic.BoolQuery

	if field[0] == '-' {
		n = n[0:]
		qq = b.MustNot
	} else {
		n = field
		qq = b.Filter
	}

	switch field {
	case "airdate":
		parseDateFilter("date", values, qq)
	case "tag":
		andFilter(n, values, qq)
	case "game_platform":
		orFilter(n, values, qq)
	case "type":
		orFilter(n, values, qq)
	default:
		logger.Info("unknown field", zap.String("field", field))
		// omit filter we don't know
	}
}

// "A" or "B" or "C" or "D".
func orFilter(field string, values []string, q func(filters ...elastic.Query) *elastic.BoolQuery) {
	q(elastic.NewTermsQuery(field, toInterface(values)...))
}

// "A" and "B" and "C" and "D".
func andFilter(field string, values []string, q func(filters ...elastic.Query) *elastic.BoolQuery) {
	for _, s := range values {
		q(elastic.NewTermQuery(field, s))
	}
}

// parse date filter like `<2020-01-20`, `>=2020-01-23`.
func parseDateFilter(field string, filters []string, q func(filters ...elastic.Query) *elastic.BoolQuery) {
	for _, s := range filters {
		switch {
		case strings.HasPrefix(s, ">="):
			q(NewRangeQuery(field).Gte(s[2:]))
		case strings.HasPrefix(s, ">"):
			q(NewRangeQuery(field).Gt(s[1:]))
		case strings.HasPrefix(s, "<="):
			q(NewRangeQuery(field).Lte(s[2:]))
		case strings.HasPrefix(s, "<"):
			q(NewRangeQuery(field).Lt(s[1:]))
		default:
			q(elastic.NewTermQuery(field, filters))
		}
	}
}

func toInterface(values []string) []interface{} {
	var f = make([]interface{}, len(values))
	for i, value := range values {
		f[i] = value
	}

	return f
}
