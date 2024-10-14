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

package tag_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/tag"
)

func getCacheRepo(t *testing.T) tag.CachedRepo {
	t.Helper()

	var r tag.CachedRepo

	test.Fx(t, fx.Provide(tag.NewCachedRepo, tag.NewMysqlRepo), fx.Populate(&r))

	return r
}

func TestCacheGet(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql, test.EnvRedis)
	t.Parallel()

	repo := getCacheRepo(t)

	_, err := repo.Get(context.Background(), 8)
	require.NoError(t, err)
}

func TestCacheGetTags(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql, test.EnvRedis)
	t.Parallel()

	repo := getCacheRepo(t)

	_, err := repo.GetByIDs(context.Background(), []model.SubjectID{1, 2, 8})
	require.NoError(t, err)
}
