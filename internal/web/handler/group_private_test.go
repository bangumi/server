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

package handler_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/internal/web/res"
)

// $ task test -- -run TestHandler_GetGroupByNamePrivate
func TestHandler_GetGroupByNamePrivate(t *testing.T) {
	t.Parallel()
	const gid = model.GroupID(5)

	u := mocks.NewUserRepo(t)
	u.EXPECT().GetByIDs(mock.Anything, model.UserID(3)).Return(map[model.UserID]model.User{
		3: {UserName: "nn", ID: 1},
	}, nil)

	g := mocks.NewGroupRepo(t)
	g.EXPECT().GetByName(mock.Anything, "g").Return(model.Group{Name: "g", ID: gid}, nil)
	g.EXPECT().ListMembersByID(mock.Anything, gid, domain.GroupMemberAll, mock.Anything, mock.Anything).
		Return([]model.GroupMember{{
			JoinAt: time.Now(),
			UserID: 3,
			Mod:    false,
		}}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: u, GroupRepo: g})

	var r res.PrivateGroupProfile
	test.New(t).Get("/p/group/g").
		Execute(app).
		JSON(&r).
		ExpectCode(http.StatusOK)
	require.Equal(t, gid, r.ID)
}

// $ task test -- -run 'TestHandler_ListGroupMembersPrivate$'
func TestHandler_ListGroupMembersPrivate(t *testing.T) {
	t.Parallel()
	const gid = model.GroupID(1)

	u := mocks.NewUserRepo(t)
	u.EXPECT().GetByIDs(mock.Anything, mock.Anything, mock.Anything).Return(map[model.UserID]model.User{
		1: {UserName: "nn", ID: 1},
	}, nil)

	g := mocks.NewGroupRepo(t)
	g.EXPECT().GetByName(mock.Anything, "g").Return(model.Group{Name: "g", ID: gid}, nil)
	g.EXPECT().CountMembersByID(mock.Anything, gid, domain.GroupMemberMod).Return(5, nil)
	g.EXPECT().ListMembersByID(mock.Anything, gid, domain.GroupMemberMod, mock.Anything, mock.Anything).
		Return([]model.GroupMember{{
			JoinAt: time.Now(),
			UserID: 1,
			Mod:    false,
		}}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: u, GroupRepo: g})

	var r test.PagedResponse
	test.New(t).Get("/p/group/g/members").
		Query("limit", "1").Query("type", "mod").
		Execute(app).
		JSON(&r).
		ExpectCode(http.StatusOK)
	require.EqualValues(t, 5, r.Total)

	var data []res.PrivateGroupMember
	require.NoError(t, json.Unmarshal(r.Data, &data))
}

// $ task test -- -run 'TestHandler_ListGroupMembersPrivate_not_found'
func TestHandler_ListGroupMembersPrivate_not_found(t *testing.T) {
	t.Parallel()

	g := mocks.NewGroupRepo(t)
	g.EXPECT().GetByName(mock.Anything, "g").Return(model.Group{}, domain.ErrNotFound)

	app := test.GetWebApp(t, test.Mock{GroupRepo: g})

	test.New(t).Get("/p/group/g/members").
		Execute(app).
		ExpectCode(http.StatusNotFound)
}

// $ task test -- -run 'TestHandler_ListGroupMembersPrivate_bad_request'
func TestHandler_ListGroupMembersPrivate_bad_request(t *testing.T) {
	t.Parallel()

	t.Run("type", func(t *testing.T) {
		t.Parallel()
		app := test.GetWebApp(t, test.Mock{})
		test.New(t).Get("/p/group/g/members").
			Query("type", "no").
			Execute(app).
			ExpectCode(http.StatusBadRequest)
	})
}
