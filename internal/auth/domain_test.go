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

package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/auth"
)

func TestAllowNsfw(t *testing.T) {
	t.Parallel()

	reg, err := time.Parse("2006-01-02", "2006-01-02")
	require.NoError(t, err)
	u := auth.Auth{
		RegTime: reg,
		ID:      1,
	}

	require.True(t, u.AllowNSFW())
}

func TestNotAllowNsfw(t *testing.T) {
	t.Parallel()

	u := auth.Auth{
		RegTime: time.Now(),
		ID:      0,
	}

	require.False(t, u.AllowNSFW())
}
