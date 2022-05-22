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

package compat

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/pkg/wiki"
)

func TestCompat_v0wiki(t *testing.T) {
	t.Parallel()
	w := wiki.Wiki{
		Type: "Crt",
		Fields: []wiki.Field{
			{Key: "简体中文名", Value: "水树奈奈"},
			{Key: "别名", Array: true, Values: []wiki.Item{
				{Key: "罗马字", Value: "Mizuki Nana"},
				{Key: "昵称", Value: "奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド"},
			}},
			{Key: "简体中文名", Value: "水树奈奈"},
			{Key: "别名", Array: true, Values: []wiki.Item{
				{Key: "罗马字", Value: "Mizuki Nana"},
				{Key: "昵称", Value: "奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド"},
			}},
		},
	}

	var expected = []interface{}{
		wikiValue{Key: "简体中文名", Value: "水树奈奈"},
		wikiValues{Key: "别名", Value: []kv{
			{Key: "罗马字", Value: "Mizuki Nana"},
			{Key: "昵称", Value: "奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド"},
		}},
		wikiValue{Key: "简体中文名", Value: "水树奈奈"},
		wikiValues{Key: "别名", Value: []kv{
			{Key: "罗马字", Value: "Mizuki Nana"},
			{Key: "昵称", Value: "奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド"},
		}},
	}

	require.Equal(t, expected, V0Wiki(w))
}
