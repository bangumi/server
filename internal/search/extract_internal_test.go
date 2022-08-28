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

package search

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseDateVal(t *testing.T) {
	t.Parallel()

	require.Equal(t, 0, parseDateVal(""))
	require.Equal(t, 20080120, parseDateVal("2008-01-20"))
	require.Equal(t, 21080620, parseDateVal("2108-06-20"))
	require.Equal(t, 0, parseDateVal("2108-06-0"))
}
