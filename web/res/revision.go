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
	"reflect"
	"time"

	"github.com/goccy/go-json"

	"github.com/bangumi/server/internal/errgo"
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

func (e *Extra) UnmarshalJSON(data []byte) error {
	var res interface{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return errgo.Wrap(err, err.Error())
	}
	typ := reflect.TypeOf(res).Kind()
	switch typ {
	case reflect.Map:
		if m, ok := res.(map[string]interface{}); ok {
			if img, ok := m["img"]; ok {
				if img, ok := img.(string); ok {
					*e = Extra{
						Img: img,
					}
				}
			}
		}
	default:
	}
	return nil
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
