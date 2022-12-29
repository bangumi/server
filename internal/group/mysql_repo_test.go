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

package group_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/group"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

const notFoundGroupID = 600
const notFoundGroupName = "not-exist-group-name"

func getRepo(t *testing.T) (group.Repo, *query.Query) {
	t.Helper()
	q := query.Use(test.GetGorm(t))
	repo, err := group.NewMysqlRepo(q, zap.NewNop())
	require.NoError(t, err)

	return repo, q
}

func prepareGroupMemberData(t *testing.T, id model.GroupID) group.Repo {
	t.Helper()
	repo, q := getRepo(t)

	test.RunAndCleanup(t, func() {
		_, err := q.WithContext(context.Background()).GroupMember.Where(q.GroupMember.GroupID.Eq(id)).Delete()
		assert.NoError(t, err)
	})

	require.NoError(t, q.WithContext(context.Background()).GroupMember.CreateInBatches([]*dao.GroupMember{
		{UserID: 1, GroupID: id, Moderator: true, CreatedTime: 1},
		{UserID: 2, GroupID: id, Moderator: false, CreatedTime: 2},
		{UserID: 3, GroupID: id, Moderator: true, CreatedTime: 3},
		{UserID: 4, GroupID: id, Moderator: true, CreatedTime: 4},
	}, 10))

	return repo
}

// $ task test-all -- -run 'TestMysqlRepo_CountMembersByName'
func TestMysqlRepo_CountMembersByName(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := prepareGroupMemberData(t, 1)

	t.Run("count all", func(t *testing.T) {
		t.Parallel()
		count, err := repo.CountMembersByID(context.Background(), 1, group.MemberAll)
		require.NoError(t, err)
		require.EqualValues(t, 4, count)
	})

	t.Run("count mod", func(t *testing.T) {
		t.Parallel()
		count, err := repo.CountMembersByID(context.Background(), 1, group.MemberMod)
		require.NoError(t, err)
		require.EqualValues(t, 3, count)
	})

	t.Run("count normal", func(t *testing.T) {
		t.Parallel()
		count, err := repo.CountMembersByID(context.Background(), 1, group.MemberNormal)
		require.NoError(t, err)
		require.EqualValues(t, 1, count)
	})
}

// $ task test-all -- -run '^TestMysqlRepo_ListMembersByName$'
func TestMysqlRepo_ListMembersByID(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	const limit = 5
	const offset = 0

	repo := prepareGroupMemberData(t, 2)

	t.Run("list all", func(t *testing.T) {
		t.Parallel()
		members, err := repo.ListMembersByID(context.Background(), 2, group.MemberAll, limit, offset)
		require.NoError(t, err)
		require.Len(t, members, 4)
		assertHaveID(t, members, 1, 2, 3, 4)
	})

	t.Run("list mod", func(t *testing.T) {
		t.Parallel()
		members, err := repo.ListMembersByID(context.Background(), 2, group.MemberMod, limit, offset)
		require.NoError(t, err)
		require.Len(t, members, 3)
		assertHaveID(t, members, 1, 3, 4)
	})

	t.Run("list normal", func(t *testing.T) {
		t.Parallel()
		members, err := repo.ListMembersByID(context.Background(), 2, group.MemberNormal, limit, offset)
		require.NoError(t, err)
		require.Len(t, members, 1)
		assertHaveID(t, members, 2)
	})

	t.Run("list offset", func(t *testing.T) {
		t.Parallel()
		members, err := repo.ListMembersByID(context.Background(), 2, group.MemberAll, limit, 1)
		require.NoError(t, err)
		require.Len(t, members, 3)
		assertHaveID(t, members, 1, 2, 3)
	})
}

func assertHaveID(t *testing.T, members []model.GroupMember, id ...model.UserID) {
	t.Helper()
	ids := make(map[model.UserID]bool)
	for _, member := range members {
		ids[member.UserID] = true
	}

	for _, userID := range id {
		require.True(t, ids[userID], fmt.Sprintf("members should contain user %d", userID))
	}
}

func TestMysqlRepo_GetByID(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()
	const groupID model.GroupID = 201
	const groupName = "group-201"

	repo, q := getRepo(t)

	test.RunAndCleanup(t, func() {
		_, err := q.WithContext(context.Background()).Group.Where(q.Group.ID.Eq(groupID)).Delete()
		assert.NoError(t, err)
	})

	err := q.WithContext(context.Background()).Group.Create(&dao.Group{
		ID:          groupID,
		Name:        groupName,
		CreatorID:   1,
		CreatedTime: uint32(time.Now().Unix()),
	})
	require.NoError(t, err)

	g, err := repo.GetByID(context.Background(), groupID)
	require.NoError(t, err)
	require.Equal(t, groupID, g.ID)
	require.Equal(t, groupName, g.Name)

	_, err = repo.GetByID(context.Background(), notFoundGroupID)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestMysqlRepo_GetByName(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()
	const groupID model.GroupID = 200
	const groupName = "group-200"

	repo, q := getRepo(t)

	test.RunAndCleanup(t, func() {
		_, err := q.WithContext(context.Background()).Group.Where(q.Group.ID.Eq(groupID)).Delete()
		assert.NoError(t, err)
	})

	err := q.WithContext(context.Background()).Group.Create(&dao.Group{
		ID:          groupID,
		Name:        groupName,
		CreatorID:   1,
		CreatedTime: uint32(time.Now().Unix()),
	})
	require.NoError(t, err)

	g, err := repo.GetByName(context.Background(), groupName)
	require.NoError(t, err)
	require.Equal(t, groupID, g.ID)
	require.Equal(t, groupName, g.Name)

	_, err = repo.GetByName(context.Background(), notFoundGroupName)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

// $ task test-all -- -run '^TestMysqlRepo_GetByName_not_found'
func TestMysqlRepo_GetByName_not_found(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()
	repo, _ := getRepo(t)

	_, err := repo.GetByName(context.Background(), t.Name())
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrNotFound)
}
