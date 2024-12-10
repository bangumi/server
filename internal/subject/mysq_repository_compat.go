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
	"github.com/trim21/go-phpserialize"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
)

type Tag struct {
	Name       *string `php:"tag_name"`
	Count      uint    `php:"result,string"`
	TotalCount uint    `php:"tag_results,string"`
}

func ParseTags(b []byte) ([]model.Tag, error) {
	var tags []Tag
	if len(b) != 0 {
		err := phpserialize.Unmarshal(b, &tags)
		if err != nil {
			return nil, errgo.Wrap(err, "ParseTags: phpserialize.Unmarshal")
		}
	}

	return slice.MapFilter(tags, func(item Tag) (model.Tag, bool) {
		if item.Name == nil {
			return model.Tag{}, false
		}
		return model.Tag{Name: *item.Name, Count: item.Count}, true
	}), nil
}
