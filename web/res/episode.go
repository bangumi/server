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

import "github.com/bangumi/server/domain"

type Episode struct {
	Airdate     string            `json:"airdate"`
	Name        string            `json:"name"`
	NameCN      string            `json:"name_cn"`
	Duration    string            `json:"duration"`
	Description string            `json:"desc"`
	Ep          float32           `json:"ep"`
	Sort        float32           `json:"sort"`
	ID          uint32            `json:"id"`
	SubjectID   uint32            `json:"subject_id"`
	Comment     uint32            `json:"comment"`
	Type        domain.EpTypeType `json:"type"`
	Disc        uint8             `json:"disc"`
}
