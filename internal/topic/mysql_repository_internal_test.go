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
	"github.com/bangumi/server/internal/model"
)

func TestMysqlRepo_convertDao(t *testing.T) {
	t.Parallel()

	p, err := convertDao(&dao.SubjectTopic{
		ID:        10,
		SubjectID: 20,
	})
	require.NoError(t, err)
	require.Equal(t, p.ID, model.TopicIDType(10))
	require.Equal(t, p.ObjectID, uint32(20))
}
