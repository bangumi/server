// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTokenScope_LegacyNull(t *testing.T) {
	t.Parallel()

	scope, legacy := parseTokenScope(sql.NullString{})
	require.True(t, legacy)
	require.Nil(t, scope)
}

func TestParseTokenScope_LegacyEmptyString(t *testing.T) {
	t.Parallel()

	scope, legacy := parseTokenScope(sql.NullString{Valid: true, String: ""})
	require.True(t, legacy)
	require.Nil(t, scope)
}

func TestParseTokenScope_Object(t *testing.T) {
	t.Parallel()

	scope, legacy := parseTokenScope(sql.NullString{Valid: true, String: `{"write:collection":true,"write:indices":false}`})
	require.False(t, legacy)
	require.Equal(t, Scope{"write:collection": true, "write:indices": false}, scope)
}

func TestParseTokenScope_NonObject(t *testing.T) {
	t.Parallel()

	scope, legacy := parseTokenScope(sql.NullString{Valid: true, String: `["write:collection"]`})
	require.False(t, legacy)
	require.Empty(t, scope)
}
