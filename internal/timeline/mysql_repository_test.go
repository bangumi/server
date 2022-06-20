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

package timeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/timeline"
)

func getRepo(t *testing.T) (domain.TimeLineRepo, *query.Query) {
	t.Helper()
	q := query.Use(test.GetGorm(t))
	repo, err := timeline.NewMysqlRepo(q, zap.NewNop())
	require.NoError(t, err)

	return repo, q
}

func Test_modelToDAO(t *testing.T) {
	var tlID model.TimeLineID = 28979826

	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()
	repo, q := getRepo(t)
	ctx := context.Background()

	tl, err := repo.GetByID(ctx, tlID)
	require.NoError(t, err)
	tlDAO, err := q.TimeLine.WithContext(ctx).Where(q.TimeLine.ID.Eq(tlID)).First()
	require.NoError(t, err)
	reDAO, err := timeline.ModelToDAO(tl)
	require.NoError(t, err)
	require.Equal(t, tlDAO, reDAO)
}
