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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/pkg/wiki"
)

func TestWiki_NonZero(t *testing.T) {
	t.Parallel()

	w := wiki.Wiki{
		Type: "t",
		Fields: []wiki.Field{
			{Key: "k", Value: "V", Values: nil, Array: false, Null: false},
			{Key: "", Value: "", Values: nil, Array: false, Null: true},
			{Key: "kk", Value: "", Values: []wiki.Item{
				{Key: "k1", Value: "v1"},
				{Key: "", Value: "v2"},
				{Key: "", Value: ""},
			}, Array: true, Null: false},
			{Key: "k", Value: "V", Values: nil, Array: false, Null: false},
			{Key: "kk", Value: "", Values: []wiki.Item{
				{Key: "k1", Value: "v1"},
				{Key: "", Value: "v2"},
				{Key: "", Value: ""},
			}, Array: true, Null: false},
		},
	}
	require.Equal(t, wiki.Wiki{
		Type: "t",
		Fields: []wiki.Field{
			{Key: "k", Value: "V", Values: nil, Array: false, Null: false},
			{Key: "kk", Value: "", Values: []wiki.Item{
				{Key: "k1", Value: "v1"},
				{Key: "", Value: "v2"},
			}, Array: true, Null: false},
			{Key: "k", Value: "V", Values: nil, Array: false, Null: false},
			{Key: "kk", Value: "", Values: []wiki.Item{
				{Key: "k1", Value: "v1"},
				{Key: "", Value: "v2"},
			}, Array: true, Null: false},
		},
	}, w.NonZero())
}
