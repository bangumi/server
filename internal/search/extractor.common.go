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

package search

import (
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/pkg/wiki"
)

func heat(s *model.Subject) uint32 {
	return s.OnHold + s.Doing + s.Dropped + s.Wish + s.Collect
}

func extractNames(s *model.Subject, w wiki.Wiki) []string {
	var names = make([]string, 0, 3)
	names = append(names, s.Name)
	if s.NameCN != "" {
		names = append(names, s.NameCN)
	}

	for _, field := range w.Fields {
		if field.Key == "别名" {
			names = append(names, getValues(field)...)
		}
	}

	return names
}

func getValues(f wiki.Field) []string {
	if f.Null {
		return nil
	}

	if !f.Array {
		return []string{f.Value}
	}

	var s = make([]string, len(f.Values))
	for i, value := range f.Values {
		s[i] = value.Value
	}
	return s
}
