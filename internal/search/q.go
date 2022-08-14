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
	"fmt"
	"strconv"
	"strings"

	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/search/syntax"
)

// keyword, filters,  error.
func parse(s string) (string, [][]string, error) {
	r, err := syntax.Parse(s)
	if err != nil {
		return "", nil, errgo.Wrap(err, "parse syntax")
	}

	var filter [][]string
	for field, values := range r.Filter {
		var op string
		if field[0] == '-' {
			op = " -= "
		} else {
			op = " = "
		}

		switch field {
		case "airdate":
			filter = append(filter, parseDateFilter(values))
		case "tag":
			for _, value := range values {
				filter = append(filter, []string{"tag" + op + value})
			}
		case "game_platform":
			filter = append(filter, values)
		case "type":
			filter = append(filter, values)
		}
	}

	return strings.Join(r.Keyword, " "), filter, nil
}

// parse date filter like `<2020-01-20`, `>=2020-01-23`.
func parseDateFilter(filters []string) []string {
	var result = make([]string, 0, len(filters))

	for _, s := range filters {
		switch {
		case strings.HasPrefix(s, ">="):
			if v, ok := parseDateValOk(s[2:]); ok {
				result = append(result, fmt.Sprintf("date >= %d", v))
			}
		case strings.HasPrefix(s, ">"):
			if v, ok := parseDateValOk(s[1:]); ok {
				result = append(result, fmt.Sprintf("date > %d", v))
			}
		case strings.HasPrefix(s, "<="):
			if v, ok := parseDateValOk(s[2:]); ok {
				result = append(result, fmt.Sprintf("date <= %d", v))
			}
		case strings.HasPrefix(s, "<"):
			if v, ok := parseDateValOk(s[1:]); ok {
				result = append(result, fmt.Sprintf("date < %d", v))
			}
		default:
			if v, ok := parseDateValOk(s); ok {
				result = append(result, fmt.Sprintf("date = %d", v))
			}
		}
	}

	return result
}

func parseDateValOk(date string) (int, bool) {
	if len(date) < 10 {
		return 0, false
	}

	// 2008-10-05 format
	if !(isDigitsOnly(date[:4]) &&
		date[4] == '-' &&
		isDigitsOnly(date[5:7]) &&
		date[7] == '-' &&
		isDigitsOnly(date[8:10])) {
		return 0, false
	}

	v, err := strconv.Atoi(date[:4])
	if err != nil {
		return 0, false
	}
	val := v * 10000

	v, err = strconv.Atoi(date[5:7])
	if err != nil {
		return 0, false
	}
	val += v * 100

	v, err = strconv.Atoi(date[8:10])
	if err != nil {
		return 0, false
	}
	val += v

	return val, true
}

func isDigitsOnly(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
