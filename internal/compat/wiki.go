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

package compat

import (
	"github.com/bangumi/server/pkg/wiki"
)

func V0Wiki(s wiki.Wiki) []interface{} {
	r := make([]interface{}, len(s.Fields))

	var valuesCount int
	for _, field := range s.Fields {
		if field.Array {
			valuesCount += len(field.Values)
		}
	}

	var kvContainer = make([]kv, 0, valuesCount)
	var lastCut, currentCut int

	for i, field := range s.Fields {
		if !field.Array {
			r[i] = wikiValue{Key: field.Key, Value: field.Value}

			continue
		}
		// non array item

		for _, value := range field.Values {
			kvContainer = append(kvContainer, kv{Key: value.Key, Value: value.Value})
		}

		currentCut += len(field.Values)
		r[i] = wikiValues{Key: field.Key, Value: kvContainer[lastCut:currentCut]}
		lastCut = currentCut
	}

	return r
}

type wikiValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type wikiValues struct {
	Key   string `json:"key"`
	Value []kv   `json:"value"`
}

type kv struct {
	Key   string `json:"k,omitempty"`
	Value string `json:"v"`
}
