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

package random_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/random"
)

func TestSecureRandomString(t *testing.T) {
	t.Parallel()
	for i := 0; i < 300; i++ {
		s := random.Base62String(32)
		require.Equal(t, 32, len(s))
	}
}

func TestBias(t *testing.T) {
	t.Parallel()
	const slen = 33
	const loop = 100000

	counts := make(map[rune]int)
	var count int64

	for i := 0; i < loop; i++ {
		s := random.Base62String(slen)
		require.Equal(t, slen, len(s))
		for _, b := range s {
			counts[b]++
			count++
		}
	}

	require.Equal(t, 62, len(counts))

	avg := float64(count) / float64(len(counts))
	for k, n := range counts {
		diff := float64(n) / avg
		if diff < 0.95 || diff > 1.05 {
			t.Errorf("Bias on '%c': expected average %f, got %d", k, avg, n)
		}
	}
}
