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

package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalPHP(t *testing.T) {
	t.Parallel()
	raw := "a:2:{s:15:\"user_wiki_apply\";s:1:\"1\";s:6:\"report\";s:1:\"1\";}"
	p, err := parseSerializedPermission([]byte(raw))
	require.NoError(t, err)
	require.True(t, p.Report)
	require.False(t, p.ManageReport)
	require.False(t, p.SubjectEdit)
}
