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

	scope, legacy := parseTokenScope(sql.NullString{
		Valid:  true,
		String: `{"write:collection":true,"write:indices":false}`,
	})
	require.False(t, legacy)
	require.Equal(t, Scope{"write:collection": true, "write:indices": false}, scope)
}

func TestParseTokenScope_OldFormats(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{`["write:collection"]`, "write:collection"} {
		scope, legacy := parseTokenScope(sql.NullString{Valid: true, String: raw})
		require.False(t, legacy)
		require.Equal(t, Scope{"write:collection": true}, scope)
	}
}

func TestParseTokenScope_ExplicitEmptyOrInvalid(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"{}", "[]", "[1]", `{"write:collection":"true"}`} {
		scope, legacy := parseTokenScope(sql.NullString{Valid: true, String: raw})
		require.False(t, legacy)
		require.NotNil(t, scope)
		require.Empty(t, scope)
	}
}
