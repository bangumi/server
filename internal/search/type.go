package search

import (
	"github.com/bangumi/server/internal/model"
)

type Record struct {
	Date   string          `json:"date"`
	Image  string          `json:"image"`
	Name   string          `json:"name"`
	NameCN string          `json:"name_cn"`
	Tags   []model.Tag     `json:"tags"`
	Score  float64         `json:"score"`
	ID     model.SubjectID `json:"id"`
	Rank   uint32          `json:"rank"`
}

type Subject struct {
	Summary      string   `json:"summary"`
	Date         string   `json:"date,omitempty"`
	Tag          []string `json:"tag,omitempty"`
	Name         []string `json:"name"`
	GamePlatform []string `json:"game_platform"`
	Record       Record   `json:"record"`
	PageRank     float64  `json:"page_rank,omitempty"`
	Heat         uint32   `json:"heat,omitempty"`
	Score        float64  `json:"score"`
	Rank         uint32   `json:"rank"`
	Platform     uint16   `json:"platform,omitempty"`
	Type         uint8    `json:"type"`
	NSFW         bool     `json:"nsfw"`
}

type Extractor = func(s model.Subject) (map[string]interface{}, error)
