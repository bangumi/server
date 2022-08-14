package search

import (
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/pkg/wiki"
)

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
			names = append(names, GetValues(field)...)
		}
	}

	return names
}

func GetValues(f wiki.Field) []string {
	if f.Null {
		return nil
	}

	if !f.Array {
		return []string{f.Value}
	}

	var s = make([]string, len(f.Values))
	for i, value := range f.Values {
		s[i] = value.Value
	}
	return s
}
