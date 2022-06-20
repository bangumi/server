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

package strparse_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/strparse"
)

func TestUserID(t *testing.T) {
	t.Parallel()
	var testCase = []struct {
		Input    string
		Expected model.UserID
	}{
		{
			Input:    "18",
			Expected: 18,
		},
		{
			Input:    "018",
			Expected: 18,
		},
	}

	for _, tc := range testCase {
		tc := tc
		t.Run(tc.Input, func(t *testing.T) {
			t.Parallel()
			u, err := strparse.UserID(tc.Input)
			require.NoError(t, err)
			require.Equal(t, tc.Expected, u)
		})
	}
}

func TestUserID_err(t *testing.T) {
	t.Parallel()

	_, err := strparse.UserID("a")
	require.Error(t, err)

	_, err = strparse.UserID("-1")
	require.Error(t, err)

	_, err = strparse.UserID("0")
	require.Error(t, err)
}
