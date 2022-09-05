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

package wiki

type Wiki struct {
	Type   string  `json:"type"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Values []Item `json:"values"`
	Array  bool   `json:"array"`
	Null   bool   `json:"null,omitempty"`
}

type Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// NonZero return a wiki without empty fields and items.
func (w Wiki) NonZero() Wiki {
	var wiki = Wiki{Type: w.Type, Fields: make([]Field, 0, len(w.Fields))}
	for _, f := range w.Fields {
		if f.Null {
			continue
		}

		if !f.Array {
			wiki.Fields = append(wiki.Fields, f)

			continue
		}

		if len(f.Values) == 0 {
			continue
		}

		var items []Item
		for _, item := range f.Values {
			if item.Value == "" {
				continue
			}

			items = append(items, item)
		}

		wiki.Fields = append(wiki.Fields, Field{
			Key:    f.Key,
			Array:  f.Array,
			Values: items,
		})
	}

	return wiki
}
