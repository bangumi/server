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

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestHandler_GetIndex_HappyPath(t *testing.T) {
	t.Parallel()
	m := mocks.NewIndexRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m})

	resp := test.New(t).Get("/v0/indices/7").Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetIndex_NSFW(t *testing.T) {
	t.Parallel()
	m := mocks.NewIndexRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7, NSFW: true}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m})

	resp := test.New(t).Get("/v0/indices/7").Execute(app)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_NewIndex_NoPermission(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(domain.AuthUserInfo{}, domain.ErrNotFound)

	app := test.GetWebApp(t, test.Mock{AuthRepo: mockAuth})

	resp := test.New(t).
		Post("/v0/indices").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]string{
			"title": "测试",
			"desc":  "测试123",
		}).Execute(app)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_NewIndex_With_Permission(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(domain.AuthUserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(domain.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().New(mock.Anything, mock.Anything).Return(nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := test.New(t).
		Post("/v0/indices").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]string{
			"title": "测试",
			"desc":  "测试123",
		}).Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_UpdateIndex_NoPermission(t *testing.T) {
	t.Parallel()
	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7, CreatorID: 6}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex})

	resp := test.New(t).
		Put("/v0/indices/7").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]string{
			"title": "测试",
			"desc":  "测试123",
		}).Execute(app)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_UpdateIndex_With_Permission(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(domain.AuthUserInfo{ID: 6, RegTime: time.Unix(1e9, 0)}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(domain.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7, CreatorID: 6}, nil)
	mockIndex.EXPECT().Update(mock.Anything, uint32(7), mock.Anything, mock.Anything).Return(nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := test.New(t).
		Put("/v0/indices/7").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]string{
			"title": "测试",
			"desc":  "测试123",
		}).Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_UpdateIndex_Invalid_Request_Data(t *testing.T) {
	t.Parallel()
	app := test.GetWebApp(t, test.Mock{})

	resp := test.New(t).
		Put("/v0/indices/7").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)

	resp = test.New(t).
		Put("/v0/indices/7").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]string{}).
		Execute(app)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
