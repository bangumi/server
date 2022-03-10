// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
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

package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/auth"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/test"
)

func getRepo(t *testing.T) domain.AuthRepo {
	t.Helper()
	repo, err := auth.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestGetByToken_NotFound(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.GetByToken(context.Background(), "not exist token")
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetByToken(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()
	repo := getRepo(t)

	u, err := repo.GetByToken(context.Background(), "a_development_access_token")
	require.NoError(t, err)

	require.Equal(t, uint32(382951), u.ID)
}

func TestGetExpired(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()
	repo := getRepo(t)

	_, err := repo.GetByToken(context.Background(), "a_expired_token")
	require.ErrorIs(t, err, domain.ErrNotFound)
}
