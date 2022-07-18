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

package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/auth"
)

func TestValue(t *testing.T) {
	t.Parallel()

	require.Equal(t, uint8(1), auth.UserGroupAdmin)
	require.Equal(t, uint8(2), auth.UserGroupBangumiAdmin)
	require.Equal(t, uint8(3), auth.UserGroupWindowAdmin)
	require.Equal(t, uint8(4), auth.UserGroupQuite)
	require.Equal(t, uint8(5), auth.UserGroupBanned)
	require.Equal(t, uint8(8), auth.UserGroupCharacterAdmin)
	require.Equal(t, uint8(9), auth.UserGroupWikiAdmin)
	require.Equal(t, uint8(10), auth.UserGroupNormal)
	require.Equal(t, uint8(11), auth.UserGroupWikiEditor)
}
