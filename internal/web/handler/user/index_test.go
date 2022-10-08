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

package user_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestHandler_BadPageParams(t *testing.T) {
	t.Parallel()

	app := test.GetWebApp(t, test.Mock{})
	var resp *test.Response

	resp = test.New(t).
		Get("/v0/users/test/indices?limit=a").
		Execute(app)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp = test.New(t).
		Get("/v0/users/test/indices?limit=1&offset=b").
		Execute(app)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp = test.New(t).
		Get("/v0/users/test/indices/collect?limit=a").
		Execute(app)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp = test.New(t).
		Get("/v0/users/test/indices/collect?limit=1&offset=b").
		Execute(app)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp = test.New(t).
		Get("/v0/users/test/indices/collect?limit=-1").
		Execute(app)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp = test.New(t).
		Get("/v0/users/test/indices/collect?offset=-1").
		Execute(app)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_UserNotFound(t *testing.T) {
	t.Parallel()

	var resp *test.Response
	mockUser := mocks.UserRepo{}

	mockUser.EXPECT().GetByName(mock.Anything, "test").Return(model.User{}, domain.ErrUserNotFound)

	app := test.GetWebApp(t, test.Mock{UserRepo: &mockUser})

	resp = test.New(t).
		Get("/v0/users/test/indices").
		Execute(app)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	resp = test.New(t).
		Get("/v0/users/test/indices/collect").
		Execute(app)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_UserExists(t *testing.T) {
	t.Parallel()

	var resp *test.Response
	mockUser := mocks.UserRepo{}

	mockUser.EXPECT().GetByName(mock.Anything, "test").Return(model.User{
		UserName: "test",
		ID:       5,
	}, nil)

	mockIndex := mocks.IndexRepo{}
	mockIndex.EXPECT().
		GetIndicesByUser(mock.Anything, model.UserID(5), mock.Anything, mock.Anything).
		Return([]model.Index{}, nil)
	mockIndex.EXPECT().
		GetCollectedIndicesByUser(mock.Anything, model.UserID(5), mock.Anything, mock.Anything).
		Return([]model.IndexCollect{}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: &mockUser, IndexRepo: &mockIndex})

	resp = test.New(t).
		Get("/v0/users/test/indices").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp = test.New(t).
		Get("/v0/users/test/indices/collect").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_CorrectPageParams(t *testing.T) {
	t.Parallel()

	var resp *test.Response
	mockUser := mocks.UserRepo{}

	mockUser.EXPECT().GetByName(mock.Anything, "test").Return(model.User{
		UserName: "test",
		ID:       5,
	}, nil)

	mockIndex := mocks.IndexRepo{}
	mockIndex.EXPECT().
		GetIndicesByUser(mock.Anything, model.UserID(5), 10, 123).
		Return([]model.Index{}, nil)
	mockIndex.EXPECT().
		GetCollectedIndicesByUser(mock.Anything, model.UserID(5), 1, 4).
		Return([]model.IndexCollect{}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: &mockUser, IndexRepo: &mockIndex})

	resp = test.New(t).
		Get("/v0/users/test/indices?limit=10&offset=123").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp = test.New(t).
		Get("/v0/users/test/indices/collect?limit=1&offset=4").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
