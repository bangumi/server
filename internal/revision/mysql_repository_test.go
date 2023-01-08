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

package revision_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/revision"
)

func getRepo(t *testing.T) revision.Repo {
	t.Helper()
	repo, err := revision.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestGetPersonRelatedBasic(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	r, err := repo.GetPersonRelated(context.Background(), 348475)
	require.NoError(t, err)
	require.EqualValues(t, 348475, r.ID)
	d, ok := r.Data["348475"]
	require.True(t, ok)
	require.Equal(t, "森岡浩之", d.Name)
}

func TestGetPersonRelatedNotFound(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.GetPersonRelated(context.Background(), 888888)
	require.Error(t, err)
}

func TestListPersonRelated(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	r, err := repo.ListPersonRelated(context.Background(), 9, 30, 0)
	require.NoError(t, err)
	require.EqualValues(t, model.UserID(181882), r[0].CreatorID)
}

func TestGetCharacterRelatedBasic(t *testing.T) {
	var rid uint32 = 1053564

	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	r, err := repo.GetCharacterRelated(context.Background(), rid)
	require.NoError(t, err)
	require.Equal(t, rid, r.ID)
	d, ok := r.Data[strconv.FormatUint(uint64(rid), 10)]
	require.True(t, ok)
	require.Equal(t, "C.C.", d.Name)
}

func TestGetCharacterRelatedNotFound(t *testing.T) {
	var rid uint32 = 888888

	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.GetCharacterRelated(context.Background(), rid)
	require.ErrorIs(t, err, gerr.ErrNotFound)
}

func TestListCharacterRelated(t *testing.T) {
	var cid model.CharacterID = 3 // character id
	var expRID uint32 = 1062064   // expected revision id

	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	r, err := repo.ListCharacterRelated(context.Background(), cid, 30, 0)
	require.NoError(t, err)
	require.Equal(t, expRID, r[0].ID)
}

func TestGetSubjectRelatedBasic(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	r, err := repo.GetSubjectRelated(context.Background(), 718391)
	require.NoError(t, err)
	require.EqualValues(t, 718391, r.ID)
	require.Equal(t, "第一次的親密接觸", r.Data.Name)
}

func TestGetSubjectRelatedNotFound(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.GetPersonRelated(context.Background(), 888888)
	require.Error(t, err)
}

func TestListSubjectRelated(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	r, err := repo.ListSubjectRelated(context.Background(), 26, 30, 0)
	require.NoError(t, err)
	require.EqualValues(t, 181882, r[0].CreatorID)
}
