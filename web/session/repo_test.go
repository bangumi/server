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

package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/gookit/goutil/timex"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/session"
)

func getRepo(t test.TB) (session.Repo, *query.Query) {
	t.Helper()
	test.RequireEnv(t, test.EnvMysql)
	q := query.Use(test.GetGorm(t))
	repo := session.NewMysqlRepo(q, logger.Copy())

	return repo, q
}

func TestMysqlRepo_Create(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	ctx := context.Background()
	r, q := getRepo(t)
	var key = "a random key " + t.Name()

	_, err := q.WithContext(ctx).WebSession.Where(q.WebSession.Key.Eq(key)).Delete()
	require.NoError(t, err)
	defer func() {
		_, err = q.WithContext(ctx).WebSession.Where(q.WebSession.Key.Eq(key)).Delete()
		require.NoError(t, err)
	}()

	err = r.Create(ctx, key, 1, session.Session{})
	require.NoError(t, err, session.ErrKeyConflict)
}

func TestMysqlRepo_Create_conflict(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	ctx := context.Background()
	r, q := getRepo(t)
	var key = "a random key " + t.Name()

	_, err := q.WithContext(ctx).WebSession.Where(q.WebSession.Key.Eq(key)).Delete()
	require.NoError(t, err)

	err = q.WithContext(ctx).WebSession.Create(&dao.WebSession{
		Key:       key,
		UserID:    1,
		Value:     []byte(`{}`),
		CreatedAt: 2,
		ExpiredAt: 1,
	})
	require.NoError(t, err)

	defer func() {
		_, err = q.WithContext(ctx).WebSession.Where(q.WebSession.Key.Eq(key)).Delete()
		require.NoError(t, err)
	}()

	err = r.Create(ctx, key, 1, session.Session{})
	require.ErrorIs(t, err, session.ErrKeyConflict)
}

func TestMysqlRepo_Get_ok(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	const uid model.IDType = 1
	ctx := context.Background()
	r, q := getRepo(t)
	var key = "a random key " + t.Name()

	_, err := q.WithContext(ctx).WebSession.Where(q.WebSession.Key.Eq(key)).Delete()
	require.NoError(t, err)

	err = q.WithContext(ctx).WebSession.Create(&dao.WebSession{
		Key:       key,
		UserID:    uid,
		Value:     []byte(`{}`),
		CreatedAt: 2,
		ExpiredAt: time.Now().Unix() + timex.OneWeekSec,
	})
	require.NoError(t, err)

	defer func() {
		_, err = q.WithContext(ctx).WebSession.Where(q.WebSession.Key.Eq(key)).Delete()
		require.NoError(t, err)
	}()

	ws, err := r.Get(ctx, key)
	require.NoError(t, err)

	require.Equal(t, key, ws.Key)
	require.Equal(t, uid, ws.UserID)
}

func TestManager_Get_notfound(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	var key = "a random key " + t.Name()
	r, _ := getRepo(t)

	_, err := r.Get(context.Background(), key)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestMysqlRepo_Revoke(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	ctx := context.Background()
	r, q := getRepo(t)
	var key = "a random key " + t.Name()

	err := q.WithContext(ctx).WebSession.Create(&dao.WebSession{Key: key, Value: []byte(`{}`)})
	require.NoError(t, err)

	defer func() {
		_, err = q.WithContext(ctx).WebSession.Where(q.WebSession.Key.Eq(key)).Delete()
		require.NoError(t, err)
	}()

	start := time.Now()
	err = r.Revoke(ctx, key)
	require.NoError(t, err)
	end := time.Now()

	s, err := q.WithContext(ctx).WebSession.Where(q.WebSession.Key.Eq(key)).First()
	require.NoError(t, err)
	require.LessOrEqual(t, start.Unix(), s.ExpiredAt)
	require.LessOrEqual(t, s.ExpiredAt, end.Unix())
}
