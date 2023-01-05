// SPDX-License-Identifier: AGPL-3.0-only
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>

package res

import (
	"time"

	"github.com/samber/lo"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/gstr"
)

const defaultShortSummaryLength = 120

type v0wiki = []any

type SubjectTag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type SubjectV0 struct {
	Date          *string               `json:"date"`
	Platform      *string               `json:"platform"`
	Image         SubjectImages         `json:"images"`
	Summary       string                `json:"summary"`
	Name          string                `json:"name"`
	NameCN        string                `json:"name_cn"`
	Tags          []SubjectTag          `json:"tags"`
	Infobox       v0wiki                `json:"infobox"`
	Rating        Rating                `json:"rating"`
	TotalEpisodes int64                 `json:"total_episodes" doc:"episodes count in database"`
	Collection    SubjectCollectionStat `json:"collection"`
	ID            model.SubjectID       `json:"id"`
	Eps           uint32                `json:"eps"`
	Volumes       uint32                `json:"volumes"`
	Redirect      model.SubjectID       `json:"-"`
	Locked        bool                  `json:"locked"`
	NSFW          bool                  `json:"nsfw"`
	TypeID        model.SubjectType     `json:"type"`
}

type SlimSubjectV0 struct {
	Date            *string           `json:"date"`
	Image           SubjectImages     `json:"images"`
	Name            string            `json:"name"`
	NameCN          string            `json:"name_cn"`
	ShortSummary    string            `json:"short_summary"`
	Tags            []SubjectTag      `json:"tags"`
	Score           float64           `json:"score"`
	Type            model.SubjectType `json:"type"`
	ID              model.SubjectID   `json:"id"`
	Eps             uint32            `json:"eps"`
	Volumes         uint32            `json:"volumes"`
	CollectionTotal uint32            `json:"collection_total"`
	Rank            uint32            `json:"rank"`
}

func ToSlimSubjectV0(s model.Subject) SlimSubjectV0 {
	var date *string
	if s.Date != "" {
		v := s.Date
		date = &v
	}
	return SlimSubjectV0{
		ID:     s.ID,
		Name:   s.Name,
		NameCN: s.NameCN,
		Date:   date,
		Tags: slice.Map(lo.Slice(s.Tags, 0, 10), func(item model.Tag) SubjectTag {
			return SubjectTag{
				Name:  item.Name,
				Count: item.Count,
			}
		}),
		ShortSummary:    gstr.Slice(s.Summary, 0, defaultShortSummaryLength),
		Image:           SubjectImage(s.Image),
		Eps:             s.Eps,
		Volumes:         s.Volumes,
		CollectionTotal: s.Collect + s.Doing + s.OnHold + s.Dropped + s.Wish,
		Rank:            s.Rating.Rank,
		Score:           s.Rating.Score,
		Type:            s.TypeID,
	}
}

type SubjectCollectionStat struct {
	OnHold  uint32 `json:"on_hold"`
	Dropped uint32 `json:"dropped"`
	Wish    uint32 `json:"wish"`
	Collect uint32 `json:"collect"`
	Doing   uint32 `json:"doing"`
}

func (s SubjectCollectionStat) Sum() uint32 {
	return s.OnHold + s.Dropped + s.Wish + s.Collect + s.Doing
}

type Count struct {
	Field1  uint32 `json:"1"`
	Field2  uint32 `json:"2"`
	Field3  uint32 `json:"3"`
	Field4  uint32 `json:"4"`
	Field5  uint32 `json:"5"`
	Field6  uint32 `json:"6"`
	Field7  uint32 `json:"7"`
	Field8  uint32 `json:"8"`
	Field9  uint32 `json:"9"`
	Field10 uint32 `json:"10"`
}

type Rating struct {
	Rank  uint32  `json:"rank"`
	Total uint32  `json:"total"`
	Count Count   `json:"count"`
	Score float64 `json:"score"`
}

type PersonRelatedSubject struct {
	Staff     string          `json:"staff"`
	Name      string          `json:"name"`
	NameCn    string          `json:"name_cn"`
	Image     string          `json:"image"`
	SubjectID model.SubjectID `json:"id"`
}

type PersonRelatedCharacter struct {
	Images        PersonImages      `json:"images"`
	Name          string            `json:"name"`
	SubjectName   string            `json:"subject_name"`
	SubjectNameCn string            `json:"subject_name_cn"`
	SubjectID     model.SubjectID   `json:"subject_id"`
	ID            model.CharacterID `json:"id"`
	Type          uint8             `json:"type" doc:"character type"`
}

type CharacterRelatedPerson struct {
	Images        PersonImages `json:"images"`
	Name          string
	SubjectName   string          `json:"subject_name"`
	SubjectNameCn string          `json:"subject_name_cn"`
	SubjectID     model.SubjectID `json:"subject_id"`
	ID            model.PersonID  `json:"id"`
	Type          uint8           `json:"type" doc:"person type"`
}

type CharacterRelatedSubject struct {
	Staff  string          `json:"staff"`
	Name   string          `json:"name"`
	NameCn string          `json:"name_cn"`
	Image  string          `json:"image"`
	ID     model.SubjectID `json:"id"`
}

type SubjectRelatedSubject struct {
	Images    SubjectImages     `json:"images"`
	Name      string            `json:"name"`
	NameCn    string            `json:"name_cn"`
	Relation  string            `json:"relation"`
	Type      model.SubjectType `json:"type"`
	SubjectID model.SubjectID   `json:"id"`
}

type SubjectRelatedCharacter struct {
	Images   PersonImages      `json:"images"`
	Name     string            `json:"name"`
	Relation string            `json:"relation"`
	Actors   []Actor           `json:"actors"`
	Type     uint8             `json:"type"`
	ID       model.CharacterID `json:"id"`
}

type SubjectRelatedPerson struct {
	Images   PersonImages   `json:"images"`
	Name     string         `json:"name" doc:"person name"`
	Relation string         `json:"relation"`
	Career   []string       `json:"career"`
	Type     uint8          `json:"type"`
	ID       model.PersonID `json:"id" doc:"person ID"`
}

type Actor struct {
	Images       PersonImages   `json:"images"`
	Name         string         `json:"name"`
	ShortSummary string         `json:"short_summary"`
	Career       []string       `json:"career"`
	ID           model.PersonID `json:"id"`
	Type         uint8          `json:"type"`
	Locked       bool           `json:"locked"`
}

type IndexSubjectV0 struct {
	AddedAt time.Time         `json:"added_at"`
	Date    *string           `json:"date"`
	Image   SubjectImages     `json:"images"`
	Name    string            `json:"name"`
	NameCN  string            `json:"name_cn"`
	Comment string            `json:"comment"`
	Infobox v0wiki            `json:"infobox"`
	ID      model.SubjectID   `json:"id"`
	TypeID  model.SubjectType `json:"type"`
}
