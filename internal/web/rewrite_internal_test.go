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

package web

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_rewrite_path(t *testing.T) {
	t.Parallel()

	require.Equal(t,
		"/users/-/collections/:subject_id/episodes",
		rewriteOpenapiPath("/users/-/collections/{subject_id}/episodes"),
	)

	require.Equal(t,
		"/v0/persons/:person_id/subjects",
		rewriteOpenapiPath("/v0/persons/{person_id}/subjects"),
	)
}
