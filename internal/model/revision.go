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

import (
	"time"
)

// RevisionCommon common parts in revision.
type RevisionCommon struct {
	CreatedAt time.Time
	Summary   string
	ID        uint32
	CreatorID uint32
	Type      uint8
}

type PersonRevision struct {
	Data PersonRevisionData
	RevisionCommon
}

type PersonRevisionData map[string]PersonRevisionDataItem

type PersonRevisionDataItem struct {
	Name       string     `json:"name" mapstructure:"prsn_name"`
	InfoBox    string     `json:"infobox" mapstructure:"prsn_infobox"`
	Summary    string     `json:"summary" mapstructure:"prsn_summary"`
	Profession Profession `json:"profession"`
	Extra      Extra      `json:"extra"`
}

type Profession struct {
	Writer      string `json:"writer,omitempty"`
	Producer    string `json:"producer,omitempty"`
	Mangaka     string `json:"mangaka,omitempty"`
	Artist      string `json:"artist,omitempty"`
	Seiyu       string `json:"seiyu,omitempty"`
	Illustrator string `json:"illustrator,omitempty"`
	Actor       string `json:"actor,omitempty"`
}

type Extra struct {
	Img string `json:"img,omitempty"`
}

type SubjectRevision struct {
	Data *SubjectRevisionData
	RevisionCommon
}

type SubjectRevisionData struct {
	Name         string
	NameCN       string
	VoteField    string
	FieldInfobox string
	FieldSummary string
	Platform     uint16
	TypeID       uint16
	SubjectID    SubjectIDType
	FieldEps     uint32
	Type         uint8
}

// CharacterRevision concrete revision data type.
type CharacterRevision struct {
	Data CharacterRevisionData
	RevisionCommon
}

type CharacterRevisionData map[string]CharacterRevisionDataItem

type CharacterRevisionDataItem struct {
	Name    string `json:"name" mapstructure:"crt_name"`
	InfoBox string `json:"infobox" mapstructure:"crt_infobox"`
	Summary string `json:"summary"`
	Extra   Extra  `json:"extra"`
}
