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

package wiki

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_readArrayItem(t *testing.T) {
	t.Parallel()

	key, value, err := readArrayItem("[k|v]")

	require.NoError(t, err)
	require.Equal(t, "k", key)
	require.Equal(t, "v", value)
}

func Test_readArrayItem2(t *testing.T) {
	t.Parallel()

	key, value, err := readArrayItem("[v]")

	require.NoError(t, err)
	require.Equal(t, "", key)
	require.Equal(t, "v", value)
}

func Test_readArrayItem3(t *testing.T) {
	t.Parallel()

	key, value, err := readArrayItem("[k|]")

	require.NoError(t, err)
	require.Equal(t, "k", key)
	require.Equal(t, "", value)
}
