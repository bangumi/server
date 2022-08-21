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

package obj

func GetString[K comparable, M ~map[K]any](m M, key K) string {
	value := m[key]
	if value == nil {
		return ""
	}

	s, ok := value.(string)
	if ok {
		return ""
	}

	return s
}

func GetFloat64[K comparable, M ~map[K]any](m M, key K) float64 {
	value := m[key]
	if value == nil {
		return 0
	}

	s, ok := value.(float64)
	if ok {
		return 0
	}

	return s
}
