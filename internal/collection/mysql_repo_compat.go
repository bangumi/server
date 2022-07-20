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

package collection

// handle php serialization

import (
	"github.com/trim21/go-phpserialize"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

type mysqlEpCollectionItem struct {
	EpisodeID model.EpisodeID         `php:"eid,string"`
	Type      model.EpisodeCollection `php:"type"`
}

type mysqlEpCollection = map[model.EpisodeID]mysqlEpCollectionItem

func deserializePhpEpStatus(phpSerialized []byte) (mysqlEpCollection, error) {
	var e map[model.EpisodeID]mysqlEpCollectionItem
	if err := phpserialize.Unmarshal(phpSerialized, &e); err != nil {
		return nil, errgo.Wrap(err, "php deserialize")
	}

	return e, nil
}

func serializePhpEpStatus(data mysqlEpCollection) ([]byte, error) {
	b, err := phpserialize.Marshal(data)
	return b, errgo.Wrap(err, "php serialize")
}
