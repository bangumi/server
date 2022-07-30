package search

import (
	"math"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/pkg/wiki"
)

func score(s *model.Subject) float64 {
	sf := s.Rating.Count

	var total = sf.Field1 + sf.Field2 + sf.Field3 + sf.Field4 + sf.Field5 +
		sf.Field6 + sf.Field7 + sf.Field8 + sf.Field9 + sf.Field10
	if total == 0 {
		return 0
	}
	var score = float64(1*sf.Field1+
		2*sf.Field2+
		3*sf.Field3+
		4*sf.Field4+
		5*sf.Field5+
		6*sf.Field6+
		7*sf.Field7+
		8*sf.Field8+
		9*sf.Field9+
		10*sf.Field10) / float64(total)

	return math.Round(score*10) / 10
}

func heat(s *model.Subject) uint32 {
	return s.OnHold + s.Doing + s.Dropped + s.Wish + s.Collect
}

func pageRank(s *model.Subject) uint32 {
	sf := s.Rating.Count
	var total = sf.Field1 + sf.Field2 + sf.Field3 + sf.Field4 + sf.Field5 +
		sf.Field6 + sf.Field7 + sf.Field8 + sf.Field9 + sf.Field10

	return total
}

func extractNames(s *model.Subject, w wiki.Wiki) []string {
	var names = make([]string, 0, 3)
	names = append(names, s.Name)
	if s.NameCN != "" {
		names = append(names, s.NameCN)
	}

	for _, field := range w.Fields {
		if field.Key == "别名" {
			if field.Null {
				continue
			}
			if !field.Array {
				names = append(names, field.Value)
			}

			for _, value := range field.Values {
				names = append(names, value.Value)
			}
		}
	}

	return names
}
