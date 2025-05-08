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

package infra

// handle php serialization

import (
	"github.com/trim21/errgo"

	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/serialize"
)

type mysqlEpCollectionItem struct {
	EpisodeID model.EpisodeID              `php:"eid,string" json:"eid,string"`
	Type      collection.EpisodeCollection `php:"type" json:"type"`
}

type mysqlEpCollection map[model.EpisodeID]mysqlEpCollectionItem

func deserializeEpStatus(serialized []byte) (mysqlEpCollection, error) {
	var e map[model.EpisodeID]mysqlEpCollectionItem
	if err := serialize.Decode(serialized, &e); err != nil {
		return nil, errgo.Wrap(err, "php deserialize")
	}

	return e, nil
}

func serializePhpEpStatus(data mysqlEpCollection) ([]byte, error) {
	b, err := serialize.Encode(data)
	return b, errgo.Wrap(err, "php serialize")
}

func (c mysqlEpCollection) toModel() collection.UserSubjectEpisodesCollection {
	var d = make(collection.UserSubjectEpisodesCollection, len(c))
	for key, value := range c {
		d[key] = collection.UserEpisodeCollection{
			ID:   value.EpisodeID,
			Type: value.Type,
		}
	}

	return d
}
