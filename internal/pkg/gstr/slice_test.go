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

func TestSlice(t *testing.T) {
	t.Parallel()

	var s = "关于我们社区指导原则"

	require.Equal(t, "", gstr.Slice(s, 0, 0))
	require.Equal(t, "关于", gstr.Slice(s, 0, 2))
	require.Equal(t, "关于我们社区指导原", gstr.Slice(s, 0, 9))
	require.Equal(t, s, gstr.Slice(s, 0, 10))
	require.Equal(t, s, gstr.Slice(s, 0, 11))
	require.Equal(t, s, gstr.Slice(s, 0, 20))

	var s2 = "1👩👩👩"
	require.Equal(t, "👩👩", gstr.Slice(s2, 1, 2))

	var s3 = "abcd"
	require.Equal(t, "cd", gstr.Slice(s3, 2, 4))
	require.Equal(t, "cd", gstr.Slice(s3, 2, 10))
}
