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

type Episode struct {
	Airdate     string
	Name        string
	NameCN      string
	Duration    string
	Description string
	Ep          float32
	SubjectID   SubjectID
	Sort        float32
	Comment     uint32
	ID          EpisodeID
	Type        EpType
	Disc        uint8
}

type EpType = uint8

const (
	EpTypeNormal  EpType = 0
	EpTypeSpecial EpType = 1
	EpTypeOpening EpType = 2
	EpTypeEnding  EpType = 3
	EpTypeMad     EpType = 4
	EpTypeOther   EpType = 6
)
