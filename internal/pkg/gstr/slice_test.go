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

func TestFirst(t *testing.T) {
	t.Parallel()

	var s = "关于我们社区指导原则"

	require.Equal(t, "", gstr.First(s, 0))
	require.Equal(t, "关于", gstr.First(s, 2))
	require.Equal(t, "关于我们社区指导原", gstr.First(s, 9))
	require.Equal(t, s, gstr.First(s, 10))
	require.Equal(t, s, gstr.First(s, 11))
	require.Equal(t, s, gstr.First(s, 20))

	var s2 = "👩👩👩"
	require.Equal(t, "👩👩", gstr.First(s2, 2))

	var s3 = "abcd"
	require.Equal(t, "ab", gstr.First(s3, 2))
}
