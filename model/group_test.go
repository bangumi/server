// Copyright (c) 2021-2022 Trim21 <trim21.me@gmail.com>
//
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

package model_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/model"
)

func TestValue(t *testing.T) {
	t.Parallel()

	require.Equal(t, uint8(1), model.GroupAdmin)
	require.Equal(t, uint8(2), model.GroupBangumiAdmin)
	require.Equal(t, uint8(3), model.GroupWindowAdmin)
	require.Equal(t, uint8(4), model.GroupQuite)
	require.Equal(t, uint8(5), model.GroupBanned)
	require.Equal(t, uint8(8), model.GroupCharacterAdmin)
	require.Equal(t, uint8(9), model.GroupWikiAdmin)
	require.Equal(t, uint8(10), model.GroupNormal)
	require.Equal(t, uint8(11), model.GroupWiki)
}
