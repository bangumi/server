// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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

	"github.com/bangumi/server/compat"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/pkg/wiki"
)

type v0wiki = []interface{}

type SubjectV0 struct {
	Date          *string               `json:"date"`
	Platform      *string               `json:"platform"`
	Image         model.SubjectImages   `json:"images"`
	Summary       string                `json:"summary"`
	Name          string                `json:"name"`
	NameCN        string                `json:"name_cn"`
	Tags          []compat.Tag          `json:"tags"`
	Infobox       v0wiki                `json:"infobox"`
	Rating        Rating                `json:"rating"`
	TotalEpisodes int64                 `json:"total_episodes"`
	Collection    SubjectCollectionStat `json:"collection"`
	ID            uint32                `json:"id"`
	Eps           uint32                `json:"eps"`
	Volumes       uint32                `json:"volumes"`
	Redirect      uint32                `json:"-"`
	Locked        bool                  `json:"locked"`
	NSFW          bool                  `json:"nsfw"`
	TypeID        model.SubjectType     `json:"type"`
}

type Subject struct {
	Image        model.SubjectImages   `json:"images"`
	Infobox      string                `json:"infobox"`
	Name         string                `json:"name"`
	NameCN       string                `json:"name_cn"`
	Summary      string                `json:"summary"`
	PlatformText string                `json:"platform_text"`
	TypeText     string                `json:"type_text"`
	Wiki         wiki.Wiki             `json:"wiki"`
	Tags         []compat.Tag          `json:"tags"`
	Rating       Rating                `json:"rating"`
	Collection   SubjectCollectionStat `json:"collection"`
	Volumes      uint32                `json:"volumes"`
	Eps          uint32                `json:"eps"`
	ID           uint32                `json:"id"`
	Redirect     uint32                `json:"-"`
	Platform     uint16                `json:"platform_id"`
	Airtime      uint8                 `json:"air_time"`
	Locked       bool                  `json:"locked"`
	NSFW         bool                  `json:"nsfw"`
	TypeID       model.SubjectType     `json:"type_id"`
}

type SubjectCollectionStat struct {
	OnHold  uint32 `json:"on_hold"`
	Dropped uint32 `json:"dropped"`
	Wish    uint32 `json:"wish"`
	Collect uint32 `json:"collect"`
	Doing   uint32 `json:"doing"`
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
	Rank  int32   `json:"rank"`
	Total uint32  `json:"total"`
	Count Count   `json:"count"`
	Score float64 `json:"score"`
}

type PersonRelatedSubject struct {
	Staff     string               `json:"staff"`
	Name      string               `json:"name"`
	NameCn    string               `json:"name_cn"`
	Image     string               `json:"image"`
	SubjectID domain.SubjectIDType `json:"id"`
}

type PersonRelatedCharacter struct {
	Images        model.PersonImages `json:"images"`
	Name          string
	SubjectName   string                 `json:"subject_name"`
	SubjectNameCn string                 `json:"subject_name_cn"`
	SubjectID     domain.SubjectIDType   `json:"subject_id"`
	ID            domain.CharacterIDType `json:"id"`
	Type          uint8                  `json:"type"`
}

type CharacterRelatedPerson struct {
	Images        model.PersonImages `json:"images"`
	Name          string
	SubjectName   string                 `json:"subject_name"`
	SubjectNameCn string                 `json:"subject_name_cn"`
	SubjectID     domain.SubjectIDType   `json:"subject_id"`
	ID            domain.CharacterIDType `json:"id"`
	Type          uint8                  `json:"type"`
}

type CharacterRelatedSubject struct {
	Staff  string               `json:"staff"`
	Name   string               `json:"name"`
	NameCn string               `json:"name_cn"`
	Image  string               `json:"image"`
	ID     domain.SubjectIDType `json:"id"`
}

type SubjectRelatedSubject struct {
	Images    model.SubjectImages  `json:"images"`
	Name      string               `json:"name"`
	NameCn    string               `json:"name_cn"`
	Relation  string               `json:"relation"`
	Type      model.SubjectType    `json:"type"`
	SubjectID domain.SubjectIDType `json:"id"`
}

type SubjectRelatedCharacter struct {
	Images   model.PersonImages  `json:"images"`
	Name     string              `json:"name"`
	Relation string              `json:"relation"`
	Actors   []Actor             `json:"actors"`
	Type     uint8               `json:"type"`
	ID       domain.PersonIDType `json:"id"`
}

type SubjectRelatedPerson struct {
	Images   model.PersonImages  `json:"images"`
	Name     string              `json:"name"`
	Relation string              `json:"relation"`
	Career   []string            `json:"career"`
	Type     uint8              `json:"type"`
	ID       domain.PersonIDType `json:"id"`
}

type Actor struct {
	Images       model.PersonImages  `json:"images"`
	Name         string              `json:"name"`
	ShortSummary string              `json:"short_summary"`
	Career       []string            `json:"career"`
	ID           domain.PersonIDType `json:"id"`
	Type         uint8               `json:"type"`
	Locked       bool                `json:"locked"`
}

type SlimSubjectV0 struct {
	AddedAt time.Time           `json:"added_at"`
	Date    *string             `json:"date"`
	Image   model.SubjectImages `json:"images"`
	Name    string              `json:"name"`
	NameCN  string              `json:"name_cn"`
	Comment string              `json:"comment"`
	Infobox v0wiki              `json:"infobox"`
	ID      uint32              `json:"id"`
	TypeID  model.SubjectType   `json:"type"`
}
