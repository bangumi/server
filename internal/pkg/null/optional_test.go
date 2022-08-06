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

package null_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/null"
)

func TestNilUint8(t *testing.T) {
	t.Parallel()

	require.Nil(t, null.NilUint8(0))

	require.Equal(t, uint8(3), *null.NilUint8(3))
}

func TestNilUint16(t *testing.T) {
	t.Parallel()

	require.Nil(t, null.NilUint16(0))

	require.Equal(t, uint16(3), *null.NilUint16(3))
}

func TestNilString(t *testing.T) {
	t.Parallel()

	require.Nil(t, null.NilString(""))

	require.Equal(t, "s", *null.NilString("s"))
}
