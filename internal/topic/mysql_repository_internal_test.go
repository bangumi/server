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

package topic

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
			in: &dao.GroupTopic{ID: 1},
			id: 1,
		},
		{
			in: &dao.SubjectTopic{ID: 2},
			id: 2,
		},
	}
	for _, tt := range tests {
		p, err := convertDao(tt.in)
		require.NoError(t, err)
		require.Equal(t, p.ID, tt.id)
	}
}
