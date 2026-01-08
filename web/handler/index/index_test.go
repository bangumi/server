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

package index_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestHandler_GetIndex_HappyPath(t *testing.T) {
	t.Parallel()
	m := mocks.NewIndexRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m})

	resp := htest.New(t, app).Get("/v0/indices/7")

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetIndex_NSFW(t *testing.T) {
	t.Parallel()
	m := mocks.NewIndexRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7, NSFW: true}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m})

	resp := htest.New(t, app).Get("/v0/indices/7")

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_GetIndex_PrivateForOwner(t *testing.T) {
	t.Parallel()
	m := mocks.NewIndexRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(
		model.Index{ID: 7, CreatorID: 6, Privacy: model.IndexPrivacyPrivate},
		nil,
	)

	mAuth := mocks.NewAuthRepo(t)
	mAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6, RegTime: time.Unix(1e9, 0)}, nil)
	mAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m, AuthRepo: mAuth})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		Get("/v0/indices/7")

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetIndex_PrivateForOthers(t *testing.T) {
	t.Parallel()
	m := mocks.NewIndexRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(
		model.Index{ID: 7, CreatorID: 6, Privacy: model.IndexPrivacyPrivate},
		nil,
	)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m})

	resp := htest.New(t, app).Get("/v0/indices/7")

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_GetIndex_Deleted(t *testing.T) {
	t.Parallel()
	m := mocks.NewIndexRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(
		model.Index{ID: 7, CreatorID: 6, Privacy: model.IndexPrivacyDeleted},
		nil,
	)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m})

	resp := htest.New(t, app).Get("/v0/indices/7")

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_NewIndex_NoPermission(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{}, gerr.ErrNotFound)

	app := test.GetWebApp(t, test.Mock{AuthRepo: mockAuth})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		BodyJSON(map[string]string{
			"title":       "测试",
			"description": "测试123",
		}).
		Post("/v0/indices")

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_NewIndex_With_Permission(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().New(mock.Anything, mock.Anything).Return(nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		BodyJSON(map[string]string{
			"title":       "测试",
			"description": "测试123",
		}).
		Post("/v0/indices")

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_UpdateIndex_NoPermission(t *testing.T) {
	t.Parallel()
	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7, CreatorID: 6}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		BodyJSON(map[string]string{
			"title":       "测试",
			"description": "测试123",
		}).
		Put("/v0/indices/7")

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_UpdateIndex_With_Permission(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6, RegTime: time.Unix(1e9, 0)}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7, CreatorID: 6}, nil)
	mockIndex.EXPECT().Update(mock.Anything, uint32(7), mock.Anything, mock.Anything).Return(nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		BodyJSON(map[string]string{
			"title":       "测试",
			"description": "测试123",
		}).
		Put("/v0/indices/7")

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_UpdateIndex_Invalid_Request_Data(t *testing.T) {
	t.Parallel()
	app := test.GetWebApp(t, test.Mock{})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		Put("/v0/indices/7")

	require.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)

	resp = htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		BodyJSON(map[string]string{}).
		Put("/v0/indices/7")

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_UpdateIndex_NonExists(t *testing.T) {
	t.Parallel()
	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{}, gerr.ErrNotFound)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		BodyJSON(map[string]string{
			"title":       "测试",
			"description": "测试123",
		}).
		Put("/v0/indices/7")

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_New_Index_Invalid_Input(t *testing.T) {
	t.Parallel()
	app := test.GetWebApp(t, test.Mock{})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		BodyJSON(map[string]string{
			"title":       "测试\001测试",
			"description": "测试123",
		}).
		Post("/v0/indices")

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Update_Index_Invalid_Input(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6, RegTime: time.Unix(1e9, 0)}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	app := test.GetWebApp(t, test.Mock{AuthRepo: mockAuth})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		BodyJSON(map[string]string{
			"title":       "测试\000",
			"description": "测试123",
		}).
		Put("/v0/indices/7")

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
