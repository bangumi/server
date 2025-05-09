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

package subject

import (
	"github.com/trim21/errgo"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/serialize"
)

type Tag struct {
	Name       *string `php:"tag_name" json:"tag_name"`
	Count      uint    `php:"result,string" json:"result,string"`
	TotalCount uint    `php:"tag_results,string" json:"tag_results,string"`
}

func ParseTags(b []byte) ([]model.Tag, error) {
	var tags []Tag
	if len(b) != 0 {
		err := serialize.Decode(b, &tags)
		if err != nil {
			return nil, errgo.Wrap(err, "ParseTags: serialize.Decode")
		}
	}

	return slice.MapFilter(tags, func(item Tag) (model.Tag, bool) {
		if item.Name == nil {
			return model.Tag{}, false
		}
		return model.Tag{Name: *item.Name, Count: item.Count, TotalCount: item.TotalCount}, true
	}), nil
}
