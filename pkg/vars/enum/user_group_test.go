// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
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

package enum_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/pkg/vars/enum"
)

func TestValue(t *testing.T) {
	t.Parallel()

	require.Equal(t, uint8(1), enum.GroupAdmin)
	require.Equal(t, uint8(2), enum.GroupBangumiAdmin)
	require.Equal(t, uint8(3), enum.GroupWindowAdmin)
	require.Equal(t, uint8(4), enum.GroupQuite)
	require.Equal(t, uint8(5), enum.GroupBanned)
	require.Equal(t, uint8(8), enum.GroupCharacterAdmin)
	require.Equal(t, uint8(9), enum.GroupWikiAdmin)
	require.Equal(t, uint8(10), enum.GroupNormal)
	require.Equal(t, uint8(11), enum.GroupWiki)
}
