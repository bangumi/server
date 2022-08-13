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

package canal

import (
	"sort"
	"strings"

	"github.com/goccy/go-json"

	"github.com/bangumi/server/internal/model"
)

func OnSubjectChange(key json.RawMessage, payload payload) {
	switch payload.Op {
	case opCreate:
	case opReplace:
		// fmt.Println(payload.After)
	case opDelete:
		// fmt.Println(payload.Before)
	case opUpdate:
		var diff = make([]string, 0, len(payload.After))
		for key, value := range payload.Before {
			if string(payload.After[key]) != string(value) {
				diff = append(diff, key)
			}
		}
		sort.Slice(diff, func(i, j int) bool {
			return strings.Compare(diff[i], diff[j]) > 0
		})
	}
}

type SubjectPayload struct {
	ID model.SubjectID `json:"subject_id"`
}
