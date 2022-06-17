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

package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// RunAndCleanup will run a function immediately and add it to t.Cleanup.
func RunAndCleanup(tb testing.TB, f func()) {
	f()
	tb.Cleanup(f)
}

// RunAndCleanupE will run a function immediately and add it to t.Cleanup.
// will also assert function return no error.
func RunAndCleanupE(tb testing.TB, f func() error) {
	require.NoError(tb, f())
	tb.Cleanup(func() {
		require.NoError(tb, f())
	})
}
