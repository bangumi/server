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
	"github.com/bangumi/server/compat"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/pkg/wiki"
)

type SubjectV0 struct {
	ID            uint32              `json:"id"`
	TypeID        model.SubjectType   `json:"type"`
	Name          string              `json:"name"`
	NameCN        string              `json:"name_cn"`
	Summary       string              `json:"summary"`
	NSFW          bool                `json:"nsfw"`
	Locked        bool                `json:"locked"`
	Date          *string             `json:"date"`
	Platform      *string             `json:"platform"`
	Image         model.SubjectImages `json:"images"`
	Infobox       []interface{}       `json:"infobox"`
	Volumes       uint32              `json:"volumes"`
	Eps           uint32              `json:"eps"`
	TotalEpisodes uint32              `json:"total_episodes"`
	Rating        Rating              `json:"rating"`
	Collection    Collection          `json:"collection"`
	Tags          []compat.Tag        `json:"tags"`
	Redirect      uint32              `json:"-"` // http 302 response
}

type Subject struct {
	ID           uint32              `json:"id"`
	Name         string              `json:"name"`
	NameCN       string              `json:"name_cn"`
	Summary      string              `json:"summary"`
	Image        model.SubjectImages `json:"images"`
	Tags         []compat.Tag        `json:"tags"`
	TypeID       model.SubjectType   `json:"type_id"`
	TypeText     string              `json:"type_text"`
	Wiki         wiki.Wiki           `json:"wiki"`
	Infobox      string              `json:"infobox"`
	Volumes      uint32              `json:"volumes"`
	Collection   Collection          `json:"collection"`
	Eps          uint32              `json:"eps"`
	Platform     uint16              `json:"platform_id"`
	PlatformText string              `json:"platform_text"`
	Airtime      uint8               `json:"air_time"`
	Locked       bool                `json:"locked"`
	NSFW         bool                `json:"nsfw"`
	Rating       Rating              `json:"rating"`
	Redirect     uint32              `json:"-"` // http 302 response
}

type Collection struct {
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
