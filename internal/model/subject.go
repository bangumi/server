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

package model

const subjectLocked = 2

type Tag struct {
	Name  string
	Count int
}

type Subject struct {
	Image         string
	Summary       string
	Name          string
	Date          string // first release date
	NameCN        string
	Infobox       string
	CompatRawTags []byte // compat field for old tags
	Tags          []Tag
	OnHold        uint32
	Dropped       uint32
	Volumes       uint32
	Eps           uint32
	Wish          uint32
	Collect       uint32
	Doing         uint32
	ID            SubjectID
	PlatformID    uint16
	TypeID        SubjectType
	Ban           uint8
	Airtime       uint8 // air weekday, start from
	NSFW          bool
	Rating        Rating
	Redirect      SubjectID
}

func (s Subject) GetID() SubjectID {
	return s.ID
}

func (s Subject) Locked() bool {
	return s.Ban == subjectLocked
}

type Count struct {
	Field1  uint32
	Field2  uint32
	Field3  uint32
	Field4  uint32
	Field5  uint32
	Field6  uint32
	Field7  uint32
	Field8  uint32
	Field9  uint32
	Field10 uint32
}

type Rating struct {
	Rank  uint32
	Total uint32
	Count Count
	Score float64
}

type Platform struct {
	Alias        string `json:"alias"`
	Type         string `json:"type"`
	TypeCN       string `json:"type_cn"`
	WikiTpl      string `json:"wiki_tpl"`
	SearchString string `json:"search_string"`
	ID           int    `json:"id"`
	EnableHeader bool   `json:"enable_header"`
}

func (p Platform) String() string {
	if p.TypeCN != "" {
		return p.TypeCN
	}

	return p.Type
}
