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

package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalPHP(t *testing.T) {
	t.Parallel()
	raw := "a:2:{s:15:\"user_wiki_apply\";s:1:\"1\";s:6:\"report\";s:1:\"1\";}"
	p, err := parseSerializedPermission([]byte(raw))
	require.NoError(t, err)
	require.True(t, p.Report)
	require.False(t, p.ManageReport)
	require.False(t, p.SubjectEdit)
}

func TestPermissionOverrides(t *testing.T) {
	t.Parallel()
	group := Permission{BanPost: true, BanVisit: true, EpEdit: true}
	for _, acl := range []string{
		`{"ban_post":false,"ban_subject_interest":true}`,
		`{"ban_post":"0","ban_subject_interest":"1"}`,
		`a:2:{s:8:"ban_post";s:1:"0";s:20:"ban_subject_interest";s:1:"1";}`,
	} {
		personal, err := decodePermission([]byte(acl))
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"ban_post", "ban_subject_interest"}, personal.DefinedFields)
		require.Equal(t, Permission{BanVisit: true, BanSubjectInterest: true, EpEdit: true}, group.Merge(personal))
		require.True(t, group.BanPost)
	}

	personal, err := decodePermission([]byte("[]"))
	require.NoError(t, err)
	require.NotNil(t, personal.DefinedFields)
	require.Empty(t, personal.DefinedFields)
	require.Equal(t, group, group.Merge(personal))
}

func TestPermissionMergePlainValues(t *testing.T) {
	t.Parallel()
	base := Permission{UserList: true, ManageReport: true, Report: true}
	other := Permission{UserList: true, BanPost: true}
	require.Equal(t, Permission{UserList: true, BanPost: true}, base.Merge(other))
}
