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

// extract game field.
func gamePlatform(s *model.Subject, w wiki.Wiki) []string {
	if s.TypeID != model.SubjectTypeGame {
		return nil
	}

	for _, field := range w.Fields {
		if field.Null {
			continue
		}
		if field.Key == "平台" {
			return GetValues(field)
		}
	}

	return nil
}
