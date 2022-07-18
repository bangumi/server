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

package subject

import (
	"fmt"
	"os"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
)

func TestParseTags(t *testing.T) {
	t.Parallel()
	type testData struct {
		Name      string `json:"name"`
		FieldTags string `json:"field_tags"`
	}

	raw, err := os.ReadFile("testdata/fields.json")
	require.NoError(t, err)

	var testCases []testData
	require.NoError(t, json.Unmarshal(raw, &testCases))

	for i, tc := range testCases {
		tags, err := parseTags([]byte(tc.FieldTags))
		require.NoError(t, err)

		require.Truef(t, len(tags) > 0, "should parse tags")

		for j, tag := range tags {
			require.NotZero(t, tag.Name, fmt.Sprintf("test case index %d[%d] '%v'", i+1, j, tag))
			require.NotZero(t, tag.Count, fmt.Sprintf("test case index %d[%d] '%v'", i+1, j, tag))
		}
	}
}
