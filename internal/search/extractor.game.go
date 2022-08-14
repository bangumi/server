package search

import (
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/pkg/wiki"
)

// extract game field.
func gamePlatform(s *model.Subject, w wiki.Wiki) (p []string) {
	if s.TypeID != model.SubjectTypeGame {
		return nil
	}

	for _, field := range w.Fields {
		if field.Null {
			continue
		}
		if field.Key == "平台" {
			return GetValues(field)
		}
	}

	return nil
}
