// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
//
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
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/revision"
	"github.com/bangumi/server/web/handler"
)

func getRepo(t *testing.T) domain.RevisionRepo {
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
	require.Equal(t, uint32(348475), r.ID)
	data, ok := r.Data.(map[string]interface{})
	require.True(t, ok)
	dat := handler.CastPersonData(data)
	d, ok := dat["348475"]
	require.True(t, ok)
	require.Equal(t, d.Name, "森岡浩之")
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
	require.Equal(t, uint32(181882), r[0].CreatorID)
}

func TestGetSubjectRelatedBasic(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	r, err := repo.GetSubjectRelated(context.Background(), 718391)
	require.NoError(t, err)
	require.Equal(t, uint32(718391), r.ID)
	data, ok := r.Data.(*model.SubjectRevisionData)
	require.True(t, ok)
	require.Equal(t, data.Name, "第一次的親密接觸")
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
	require.Equal(t, uint32(181882), r[0].CreatorID)
}
