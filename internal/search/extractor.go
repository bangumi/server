package search

import (
	"strconv"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/pkg/wiki"
)

// ExtractSubject extract indexed data from db subject row.
func (c *Client) ExtractSubject(s *model.Subject) Subject {
	tags := s.Tags

	w := wiki.ParseOmitError(s.Infobox)

	rank := s.Rating.Total
	score := s.Rating.Score

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
		Date:         parseDateVal(s.Date),
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
	}
}

func parseDateVal(date string) int {
	if len(date) < 10 {
		return 0
	}

	// 2008-10-05 format
	var val = 0

	v, err := strconv.Atoi(date[:4])
	if err != nil {
		return 0
	}
	val = v * 10000

	v, err = strconv.Atoi(date[5:7])
	if err != nil {
		return 0
	}
	val += v * 100

	v, err = strconv.Atoi(date[8:10])
	if err != nil {
		return 0
	}
	val += v

	return val
}
