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

package duration

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/trim21/errgo"
)

// ErrOverFlow is the error when int components too big like `1:61:00`.
var ErrOverFlow = errors.New(
	"overflow: components like minutes or seconds is bigger than it should be",
)

func ParseOmitError(s string) time.Duration {
	d, err := Parse(s)
	if err != nil {
		return 0
	}

	return d
}

var extraPattern = regexp.MustCompile(`^(\d+(:))?(\d+):(\d+)$`)

// Parse a string like "01:31:41" and go default Duration.String() to time.Duration.
func Parse(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}

	g := extraPattern.FindStringSubmatch(s)
	if g == nil {
		d, err := time.ParseDuration(s)

		return d, errgo.Wrap(err, "time.ParseDuration")
	}

	var hour int
	var err error
	if g[1] != "" {
		hour, err = strconv.Atoi(g[1][:len(g[1])-1])
		if err != nil {
			return 0, errgo.Wrap(err, "parse time component as int")
		}
	}

	minute, err := strconv.Atoi(g[3])
	if err != nil {
		return 0, errgo.Wrap(err, "parse time component as int")
	} else if minute >= 60 {
		return 0, errgo.Msg(ErrOverFlow,
			fmt.Sprintf("overflow: minutes %d is bigger than 60", minute))
	}

	second, err := strconv.Atoi(g[4])
	if err != nil {
		return 0, errgo.Wrap(err, "parse time component as int")
	} else if second >= 60 {
		return 0, errgo.Msg(ErrOverFlow,
			fmt.Sprintf("overflow: second %d is bigger than 60", second))
	}

	return time.Hour*time.Duration(hour) +
		time.Minute*time.Duration(minute) + time.Second*time.Duration(second), nil
}
