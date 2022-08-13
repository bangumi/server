package search

import (
	"context"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/pkg/wiki"
)

// ExtractSubject extract indexed data from db subject row.
func (c *Client) ExtractSubject(ctx context.Context, s *model.Subject) (Subject, error) {
	tags := s.Tags

	w := wiki.ParseOmitError(s.Infobox)

	rank := pageRank(s)
	score := score(s)

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	return Subject{
		Name:         extractNames(s, w),
		Tag:          tagNames,
		Summary:      s.Summary,
		NSFW:         s.NSFW,
		Type:         s.TypeID,
		Date:         s.Date,
		Platform:     s.PlatformID,
		GamePlatform: gamePlatform(s, w),
		PageRank:     float64(rank),
		Rank:         s.Rating.Rank,
		Heat:         heat(s),
		Score:        score,
		Record: Record{
			ID:     s.ID,
			Image:  s.Image,
			Name:   s.Name,
			NameCN: s.NameCN,
			Date:   s.Date,
			Tags:   tags,
			Rank:   rank,
			Score:  score,
		},
	}, nil
}
