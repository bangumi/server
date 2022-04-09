// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
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
)

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

type PersonRevisionDataItem struct {
	InfoBox    string     `json:"prsn_infobox"`
	Summary    string     `json:"prsn_summary"`
	Profession Profession `json:"profession"`
	Extra      Extra      `json:"extra"`
	Name       string     `json:"prsn_name"`
}

type PersonRevision struct {
	CreatedAt time.Time                         `json:"created_at"`
	Data      map[string]PersonRevisionDataItem `json:"data"`
	Creator   Creator                           `json:"creator"`
	Summary   string                            `json:"summary"`
	ID        uint32                            `json:"id"`
	Type      uint8                             `json:"type"`
}

type SubjectRevisionData struct {
	Name         string `json:"name"`
	NameCN       string `json:"name_cn"`
	VoteField    string `json:"vote_field"`
	FieldInfobox string `json:"field_infobox"`
	FieldSummary string `json:"field_summary"`
	Platform     uint16 `json:"platform"`
	TypeID       uint16 `json:"type_id"`
	SubjectID    uint32 `json:"subject_id"`
	FieldEps     uint32 `json:"field_eps"`
	Type         uint8  `json:"type"`
}

type SubjectRevision struct {
	CreatedAt time.Time            `json:"created_at"`
	Data      *SubjectRevisionData `json:"data"`
	Creator   Creator              `json:"creator"`
	Summary   string               `json:"summary"`
	ID        uint32               `json:"id"`
	Type      uint8                `json:"type"`
}
