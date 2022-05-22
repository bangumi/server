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

package compat_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/compat"
)

type testData struct {
	Name      string `json:"name"`
	FieldTags string `json:"field_tags"`
}

func TestParseTags_WithNil(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/subject_8_tags.txt")
	require.NoError(t, err)

	tags, err := compat.ParseTags(bytes.TrimSpace(raw))
	require.NoError(t, err)

	assert.EqualValues(t, []compat.Tag{
		{Count: 1645, Name: "叛逆的鲁鲁修"},
		{Count: 1229, Name: "SUNRISE"},
		{Count: 936, Name: "反逆のルルーシュ"},
		{Count: 721, Name: "还是死妹控"},
		{Count: 664, Name: "TV"},
		{Count: 603, Name: "妹控"},
		{Count: 569, Name: "codegeass"},
		{Count: 523, Name: "谷口悟朗"},
		{Count: 453, Name: "鲁路修"},
		{Count: 427, Name: "R2"},
		{Count: 409, Name: "2008"},
		{Count: 385, Name: "原创"},
		{Count: 357, Name: "2008年4月"},
		{Count: 174, Name: "大河内一楼"},
		{Count: 151, Name: "日升"},
		{Count: 120, Name: "萝卜"},
		{Count: 111, Name: "机战"},
		{Count: 104, Name: "狗得鸡鸭死"},
		{Count: 94, Name: "福山润"},
		{Count: 84, Name: "露露胸"},
		{Count: 69, Name: "CLAMP"},
		{Count: 67, Name: "科幻"},
		{Count: 62, Name: "鲁鲁修"},
		{Count: 57, Name: "GEASS"},
		{Count: 54, Name: "神作"},
		{Count: 49, Name: "战斗"},
		{Count: 41, Name: "战争"},
		{Count: 40, Name: "裸露修的跌二次KUSO"},
		{Count: 37, Name: "中二"},
		{Count: 34, Name: "樱井孝宏"},
	}, tags)
}

func TestParseTags(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/fields.json")
	require.NoError(t, err)

	var testCases []testData
	require.NoError(t, json.Unmarshal(raw, &testCases))

	for _, tc := range testCases {
		tc := tc

		t.Run("", func(t *testing.T) {
			t.Parallel()

			tags, err := compat.ParseTags([]byte(tc.FieldTags))
			require.NoError(t, err)

			require.Truef(t, len(tags) > 0, "should parse tags")
		})
	}
}
