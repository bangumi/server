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

package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSameSimpleType(t *testing.T) {
	t.Parallel()

	require.True(t, sameSimpleType(reflect.TypeOf((*int)(nil)), reflect.TypeOf((*int)(nil))))
	require.True(t, sameSimpleType(reflect.TypeOf(int(1)), reflect.TypeOf(int(1))))
	require.True(t, sameSimpleType(reflect.TypeOf("int(1)"), reflect.TypeOf("")))

	require.False(t, sameSimpleType(reflect.TypeOf((**int)(nil)), reflect.TypeOf((**int)(nil))))
	require.False(t, sameSimpleType(reflect.TypeOf(map[int]bool{}), reflect.TypeOf(map[int]bool{})))
	require.False(t, sameSimpleType(reflect.TypeOf(struct{}{}), reflect.TypeOf(struct{}{})))
}
