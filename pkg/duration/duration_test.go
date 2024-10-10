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

package duration_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/pkg/duration"
)

func TestParse(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		Err        error
		ErrPattern *regexp.Regexp
		Name       string
		Input      string
		Expected   time.Duration
	}{
		{
			Name:     "stdlib duration parser",
			Input:    "1h89m1s",
			Expected: time.Hour + time.Minute*89 + time.Second,
		},
		{
			Name:     "colon separated minute",
			Input:    "23:58",
			Expected: time.Minute*23 + time.Second*58,
		},
		{
			Name:     "colon separated hour",
			Input:    "1:02:3",
			Expected: time.Hour + time.Minute*2 + time.Second*3,
		},
		{
			Name:       `minute overflow`,
			Input:      "1:61:3",
			ErrPattern: regexp.MustCompile("overflow.*minute"),
			Err:        duration.ErrOverFlow,
		},
		{
			Name:       `second overflow`,
			Input:      "1:6:300",
			ErrPattern: regexp.MustCompile("overflow.*second"),
			Err:        duration.ErrOverFlow,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			actual, err := duration.Parse(tc.Input)
			if tc.ErrPattern != nil {
				require.NotNil(t, err, "expecting error")
				require.Truef(t, tc.ErrPattern.MatchString(err.Error()),
					"error message should match, err: '%s'", err.Error())

				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.Expected, actual, "parse result should be same as expected")
		})
	}
}
