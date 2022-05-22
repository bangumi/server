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
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/pkg/wiki"
)

const large = `  {{Infobox animanga/TVAnime
|中文名= 潜行吧！奈亚子W
|别名={
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
}
|话数= 12
|放送开始= 2013年4月7日
|放送星期= 
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|别名={
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
}
|     播放结束      = 2013年6月30日
|  播放结束       = 2013年6月30日
|                             播放结束= 2013年6月30日

|播放结束= 2013年6月30日
|别名={


[2013年6月30日]


[2013年6月30日]
[2013年6月30日]
[2013年6月30日]

[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
}
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日













|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|别名={
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[           2013年6月30日]
}
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|其他= 
|Copyright= 
|原作= 逢空万太
|导演= 長澤剛
|人物设定= 滝山真哲
}}    ` + "\t"

const medium = `  {{Infobox animanga/TVAnime
|中文名= 潜行吧！奈亚子W
|别名={
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
}
|话数= 12
|放送开始= 2013年4月7日
|放送星期= 
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|别名={
[1|2013年6月30日]
[2|2013年6月30日]
[3|2013年6月30日]
[4|2013年6月30日]
[4|2013年6月30日]
[4|2013年6月30日]
[4|2013年6月30日]
}
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|其他= 
|Copyright= 
|原作= 逢空万太
|导演= 長澤剛
|人物设定= 滝山真哲
}}    ` + "\t"

const small = `  {{Infobox animanga/TVAnime
|中文名= 潜行吧！奈亚子W
|别名={
[2013年6月30日]
}
|话数= 12
|别名={
[1|2013年6月30日]
}
|播放结束= 2013年6月30日
|其他= 
|Copyright= 
}}    ` + "\t"

func BenchmarkParse_large(b *testing.B) {
	var w wiki.Wiki
	for i := 0; i < b.N; i++ {
		w, _ = wiki.Parse(large)
	}
	runtime.KeepAlive(w)
}

func BenchmarkParse_medium(b *testing.B) {
	var w wiki.Wiki
	for i := 0; i < b.N; i++ {
		w, _ = wiki.Parse(medium)
	}
	runtime.KeepAlive(w)
}

func BenchmarkParse_small(b *testing.B) {
	var w wiki.Wiki
	for i := 0; i < b.N; i++ {
		w, _ = wiki.Parse(small)
	}
	runtime.KeepAlive(w)
}

func BenchmarkWiki_NonZero(b *testing.B) {
	var w, err = wiki.Parse(benchNonZeroInput)
	require.NoError(b, err)

	var r wiki.Wiki
	for i := 0; i < b.N; i++ {
		r = w.NonZero()
	}
	runtime.KeepAlive(r)
}

const benchNonZeroInput = `  {{Infobox animanga/TVAnime
|中文名= 潜行吧！奈亚子W
|别名={
}
|话数= 12
|放送开始= 2013年4月7日
|放送星期= 
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束=
|播放结束=
|播放结束=
|播放结束= 2013年6月30日
|别名={
[]
[2013年6月30日]
[2013年6月30日]
[]
[2013年6月30日]
[2013年6月30日]
[]
[2013年6月30日]
[2013年6月30日]
}
|     播放结束      = 2013年6月30日
|  播放结束       = 2013年6月30日
|                             播放结束= 2013年6月30日
|                             播放结束= 2013年6月30日
|                             播放结束=

|播放结束= 2013年6月30日
|别名={


[2013年6月30日]


[2013年6月30日]
[2013年6月30日]
[2013年6月30日]

[2013年6月30日]
[2013年6月30日]
[]
[2013年6月30日]
[]
}
|播放结束= 2013年6月30日
|                             播放结束=
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日













|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|                             播放结束=
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|别名={
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[2013年6月30日]
[           2013年6月30日]
}
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|播放结束= 2013年6月30日
|其他= 
|Copyright= 
|原作= 逢空万太
|导演= 長澤剛
|人物设定= 滝山真哲
}}    ` + "\t"
