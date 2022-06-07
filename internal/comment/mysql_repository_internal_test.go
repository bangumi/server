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

package comment

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/dal/dao"
)

func TestMysqlRepo_convertDao(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in interface{}
		id uint32
	}{
		{
			id: 1,
			in: &dao.SubjectTopicComment{ID: 1},
		},
		{
			id: 2,
			in: &dao.GroupTopicComment{ID: 2},
		},
		{
			id: 3,
			in: &dao.IndexComment{ID: 3},
		},
		{
			id: 4,
			in: &dao.EpisodeComment{ID: 4},
		},
		{
			id: 5,
			in: &dao.CharacterComment{ID: 5},
		},
	}
	for _, tt := range tests {
		s, err := convertDao(tt.in)
		require.NoError(t, err)
		require.Equal(t, s.ID, tt.id)
	}
}

func TestMysqlRepo_convertModelComments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  interface{}
		err error
		len int
	}{
		{
			len: 0,
			in:  nil,
			err: errInputNilComments,
		},
		{
			len: 2,
			in: []interface{}{
				&dao.CharacterComment{ID: 2},
				&dao.CharacterComment{ID: 5},
			},
		},
	}
	for _, tt := range tests {
		s, err := convertModelComments(tt.in)
		require.ErrorIs(t, err, tt.err)
		require.Equal(t, len(s), tt.len)
	}
}
