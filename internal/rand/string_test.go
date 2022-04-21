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

package rand_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/rand"
)

func TestSecureRandomString(t *testing.T) {
	s := rand.SecureRandomString(32)
	require.Equal(t, 32, len(s))
	fmt.Println(s)
}

func BenchmarkSecureRandomString(b *testing.B) {
	var s string
	for i := 0; i < b.N; i++ {
		s = rand.SecureRandomString(32)
	}

	runtime.KeepAlive(s)
}
