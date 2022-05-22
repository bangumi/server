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

package wiki_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/pkg/wiki"
)

func TestParseFull(t *testing.T) {
	t.Parallel()
	value, err := wiki.Parse(`{{Infobox Crt
|简体中文名= 水树奈奈
|官网= https://www.mizukinana.jp
|FanClub= https://fanclub.mizukinana.jp
|Twitter= https://twitter.com/NM_NANAPARTY
}}`)

	require.NoError(t, err)

	require.Equal(t, wiki.Wiki{
		Type: "Crt",
		Fields: []wiki.Field{
			{Key: "简体中文名", Value: "水树奈奈"},
			{Key: "官网", Value: "https://www.mizukinana.jp"},
			{Key: "FanClub", Value: "https://fanclub.mizukinana.jp"},
			{Key: "Twitter", Value: "https://twitter.com/NM_NANAPARTY"},
		}}, value)
}

var expected = wiki.Wiki{
	Type: "Crt",
	Fields: []wiki.Field{
		{Key: "简体中文名", Value: "水树奈奈"},
		{Key: "别名", Array: true, Values: []wiki.Item{
			{Key: "", Value: "第二中文名"},
			{Key: "", Value: "英文名"},
			{Key: "日文名", Value: "近藤奈々 (こんどう なな)"},
			{Key: "纯假名", Value: "みずき なな"},
			{Key: "罗马字", Value: "Mizuki Nana"},
			{Key: "昵称", Value: "奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド"},
			{Key: "其他名义", Value: ""},
		}},
	},
}

func TestParseFullArray(t *testing.T) {
	t.Parallel()
	value, err := wiki.Parse(`{{Infobox Crt
|简体中文名= 水树奈奈
|别名={
[第二中文名]
[英文名]
[日文名|近藤奈々 (こんどう なな)]
[纯假名|みずき なな]
[罗马字|Mizuki Nana]
[昵称|奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド]
[其他名义|]
}
}}`)

	require.NoError(t, err)

	require.Equal(t, expected, value)
}

func TestParseEmptyLine(t *testing.T) {
	t.Parallel()
	value, err := wiki.Parse(`{{Infobox Crt
|简体中文名= 水树奈奈
|别名={


[第二中文名]
[英文名]
[日文名|近藤奈々 (こんどう なな)]

[纯假名|みずき なな]
[罗马字|Mizuki Nana]
[昵称|奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド]
[其他名义|]

}
}}`)

	require.NoError(t, err)
	require.Equal(t, expected, value)
}

func TestParseExtraSpace(t *testing.T) {
	t.Parallel()
	value, err := wiki.Parse(`{{Infobox Crt
|简体中文名= 水树奈奈
| 别名 = {
[第二中文名]
[ 英文名]
[日文名|近藤奈々 (こんどう なな)]
[纯假名 |みずき なな]
[罗马字|Mizuki Nana]
[昵称|奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド]
[其他名义|]

 }
}}`)

	require.NoError(t, err)
	require.Equal(t, expected, value)
}

func TestArrayNoClose(t *testing.T) {
	t.Parallel()
	_, err := wiki.Parse(`{{Infobox Crt
| 别名 = {

[昵称|奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド]
[其他名义|]
}}`)

	require.ErrorIs(t, err, wiki.ErrWikiSyntax)
	require.Regexp(t, regexp.MustCompile("array.*close"), err)
}

func TestArrayNoClose2(t *testing.T) {
	t.Parallel()
	_, err := wiki.Parse(`{{Infobox Crt
| 别名 = {

[昵称|奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド]
[其他名义|]
|简体中文名= 水树奈奈
}}`)

	require.ErrorIs(t, err, wiki.ErrWikiSyntax)
	require.Regexp(t, regexp.MustCompile("array.*closed"), err)
	require.Regexp(t, regexp.MustCompile("line: 6"), err)
}

func TestArrayNoClose_empty_item(t *testing.T) {
	t.Parallel()
	_, err := wiki.Parse(`{{Infobox Crt
| 别名 = {
}}`)

	require.ErrorIs(t, err, wiki.ErrWikiSyntax)
	require.Regexp(t, regexp.MustCompile("array.*closed"), err)
	require.Regexp(t, regexp.MustCompile("line: 3"), err)
}

func TestScalar_No_sign_equal(t *testing.T) {
	t.Parallel()
	_, err := wiki.Parse(`{{Infobox Crt
| 别名 
}}`)

	require.ErrorIs(t, err, wiki.ErrExpectingSignEqual)
	require.Regexp(t, regexp.MustCompile("别名"), err)
	require.Regexp(t, regexp.MustCompile("line: 2"), err)
}

func TestTypeNoLineBreak(t *testing.T) {
	t.Parallel()
	w, err := wiki.Parse(`{{Infobox Crt}}`)
	require.NoError(t, err)
	require.Equal(t, "Crt", w.Type)
}
