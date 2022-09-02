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

package pm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/pm"
)

func getRepo(t *testing.T) domain.PrivateMessageRepo {
	t.Helper()
	repo, err := pm.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestListInbox(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	list, err := repo.List(context.Background(), model.UserID(1), model.PrivateMessageFolderTypeInbox, 0, 10)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestListOutbox(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	list, err := repo.List(context.Background(), model.UserID(1), model.PrivateMessageFolderTypeOutbox, 0, 10)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestListRelated(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	list, err := repo.ListRelated(context.Background(), model.UserID(1), model.PrivateMessageID(1))
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestCountTypes(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	counts, err := repo.CountTypes(context.Background(), model.UserID(1))
	require.NoError(t, err)
	require.Equal(t, counts.Inbox, 0)
	require.Equal(t, counts.Outbox, 0)
	require.Empty(t, counts.Unread, 0)
}

func TestListRecentContact(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	list, err := repo.ListRecentContact(context.Background(), model.UserID(1))
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestMarkRead(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	err := repo.MarkRead(context.Background(), model.UserID(1), model.PrivateMessageID(1))
	require.Error(t, err)
}

func TestCreate(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	res, err := repo.Create(
		context.Background(),
		model.UserID(1),
		[]model.UserID{model.UserID(382951)},
		domain.PrivateMessageIDFilter{Type: null.NewFromPtr[model.PrivateMessageID](nil)},
		"私信",
		"内容",
	)
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestDelete(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	res, err := repo.Create(
		context.Background(),
		model.UserID(1),
		[]model.UserID{model.UserID(382951)},
		domain.PrivateMessageIDFilter{Type: null.NewFromPtr[model.PrivateMessageID](nil)},
		"私信",
		"内容",
	)
	require.NoError(t, err)
	require.Len(t, res, 1)
	err = repo.Delete(context.Background(), model.UserID(1), []model.PrivateMessageID{res[0].ID})
	require.NoError(t, err)
}
