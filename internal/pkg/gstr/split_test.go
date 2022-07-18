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

package gstr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/gstr"
)

func TestPartition(t *testing.T) {
	t.Parallel()

	key, value := gstr.Partition("a=b", '=')
	require.Equal(t, "a", key)
	require.Equal(t, "b", value)

	key, value = gstr.Partition("a=", '=')
	require.Equal(t, "a", key)
	require.Equal(t, "", value)

	key, value = gstr.Partition("=", '=')
	require.Equal(t, "", key)
	require.Equal(t, "", value)

	key, value = gstr.Partition("ab", '=')
	require.Equal(t, "ab", key)
	require.Equal(t, "", value)
}

func TestSplit(t *testing.T) {
	t.Parallel()

	s := gstr.Split("a=b", "=")
	require.Equal(t, []string{"a", "b"}, s)

	s = gstr.Split("a==b", "=")
	require.Equal(t, []string{"a", "b"}, s)

	s = gstr.Split("", "=")
	require.Equal(t, []string{}, s)
}
