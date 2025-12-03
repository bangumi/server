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

	wiki "github.com/bangumi/wiki-parser-go"
	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/compat"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/tag"
	"github.com/bangumi/server/pkg/vars"
)

const defaultShortSummaryLength = 120

type V0wiki = []any

type SubjectTag struct {
	Name      string `json:"name"`
	Count     uint   `json:"count"`
	TotalCont uint   `json:"total_cont"`
}

type SubjectV0 struct {
	Date          *string               `json:"date"`
	Platform      *string               `json:"platform"`
	Images        SubjectImages         `json:"images"`
	Summary       string                `json:"summary"`
	Name          string                `json:"name"`
	NameCN        string                `json:"name_cn"`
	Tags          []SubjectTag          `json:"tags"`
	Infobox       V0wiki                `json:"infobox"`
	Rating        Rating                `json:"rating"`
	TotalEpisodes int64                 `json:"total_episodes" doc:"episodes count in database"`
	Collection    SubjectCollectionStat `json:"collection"`
	ID            model.SubjectID       `json:"id"`
	Eps           uint32                `json:"eps"`
	MetaTags      []string              `json:"meta_tags"`
	Volumes       uint32                `json:"volumes"`
	Series        bool                  `json:"series"`
	Locked        bool                  `json:"locked"`
	NSFW          bool                  `json:"nsfw"`
	TypeID        model.SubjectType     `json:"type"`
	Redirect      model.SubjectID       `json:"-"`
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

func PlatformString(s model.Subject) *string {
	platform, ok := vars.PlatformMap[s.TypeID][s.PlatformID]
	if !ok && s.TypeID != 0 {
		logger.Warn("unknown platform",
			zap.Uint32("subject", s.ID),
			zap.Uint8("type", s.TypeID),
			zap.Uint16("platform", s.PlatformID),
		)

		return nil
	}
	v := platform.String()
	return &v
}

func ToSubjectV0(s model.Subject, totalEpisode int64, metaTags []tag.Tag) SubjectV0 {
	return SubjectV0{
		TotalEpisodes: totalEpisode,
		ID:            s.ID,
		Images:        SubjectImage(s.Image),
		Summary:       s.Summary,
		Name:          s.Name,
		Platform:      PlatformString(s),
		NameCN:        s.NameCN,
		Date:          null.NilString(s.Date),
		Infobox:       compat.V0Wiki(wiki.ParseOmitError(s.Infobox).NonZero()),
		Volumes:       s.Volumes,
		Redirect:      s.Redirect,
		Eps:           s.Eps,
		MetaTags: lo.Map(metaTags, func(item tag.Tag, index int) string {
			return item.Name
		}),
		Tags: slice.Map(s.Tags, func(tag model.Tag) SubjectTag {
			return SubjectTag{
				Name:  tag.Name,
				Count: tag.Count,
			}
		}),
		Collection: SubjectCollectionStat{
			OnHold:  s.OnHold,
			Wish:    s.Wish,
			Dropped: s.Dropped,
			Collect: s.Collect,
			Doing:   s.Doing,
		},
		TypeID: s.TypeID,
		Series: s.Series,
		Locked: s.Locked(),
		NSFW:   s.NSFW,
		Rating: Rating{
			Rank:  s.Rating.Rank,
			Total: s.Rating.Total,
			Count: Count{
				Field1:  s.Rating.Count.Field1,
				Field2:  s.Rating.Count.Field2,
				Field3:  s.Rating.Count.Field3,
				Field4:  s.Rating.Count.Field4,
				Field5:  s.Rating.Count.Field5,
				Field6:  s.Rating.Count.Field6,
				Field7:  s.Rating.Count.Field7,
				Field8:  s.Rating.Count.Field8,
				Field9:  s.Rating.Count.Field9,
				Field10: s.Rating.Count.Field10,
			},
			Score: s.Rating.Score,
		},
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
	Staff     string            `json:"staff"`
	Eps       string            `json:"eps" doc:"episodes participated"`
	Name      string            `json:"name"`
	NameCn    string            `json:"name_cn"`
	Image     string            `json:"image"`
	Type      model.SubjectType `json:"type"`
	SubjectID model.SubjectID   `json:"id"`
}

type PersonRelatedCharacter struct {
	Images        PersonImages      `json:"images"`
	Name          string            `json:"name"`
	SubjectName   string            `json:"subject_name"`
	SubjectNameCn string            `json:"subject_name_cn"`
	SubjectType   model.SubjectType `json:"subject_type"`
	SubjectID     model.SubjectID   `json:"subject_id"`
	Staff         string            `json:"staff"`
	ID            model.CharacterID `json:"id"`
	Type          uint8             `json:"type" doc:"character type"`
}

type CharacterRelatedPerson struct {
	Images        PersonImages      `json:"images"`
	Name          string            `json:"name"`
	SubjectName   string            `json:"subject_name"`
	SubjectNameCn string            `json:"subject_name_cn"`
	SubjectType   model.SubjectType `json:"subject_type"`
	SubjectID     model.SubjectID   `json:"subject_id"`
	Staff         string            `json:"staff"`
	ID            model.PersonID    `json:"id"`
	Type          uint8             `json:"type" doc:"person type"`
}

type CharacterRelatedSubject struct {
	Staff  string            `json:"staff"`
	Name   string            `json:"name"`
	NameCn string            `json:"name_cn"`
	Image  string            `json:"image"`
	Type   model.SubjectType `json:"type"`
	ID     model.SubjectID   `json:"id"`
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
	Summary  string            `json:"summary"`
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
	Eps      string         `json:"eps" doc:"episodes participated"`
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
	Infobox V0wiki            `json:"infobox"`
	ID      model.SubjectID   `json:"id"`
	TypeID  model.SubjectType `json:"type"`
}
