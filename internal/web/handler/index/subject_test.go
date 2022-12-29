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

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestHandler_Add_Index_Subject(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{
		CreatorID: 6,
		ID:        7,
	}, nil)
	mockIndex.EXPECT().
		AddOrUpdateIndexSubject(mock.Anything, model.IndexID(7), model.SubjectID(5), uint32(48), "test123").
		Return(&index.Subject{}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := test.New(t).
		Post("/v0/indices/7/subjects").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]any{
			"subject_id": 5,
			"sort":       48,
			"comment":    "test123",
		}).
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_Add_Index_Subject_NoPermission(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{
		CreatorID: 1,
		ID:        7,
	}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := test.New(t).
		Post("/v0/indices/7/subjects").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]any{
			"subject_id": 5,
			"sort":       48,
			"comment":    "test123",
		}).
		Execute(app)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_Update_Index_Subject(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{
		CreatorID: 6,
		ID:        7,
	}, nil)
	mockIndex.EXPECT().
		AddOrUpdateIndexSubject(mock.Anything, model.IndexID(7), model.SubjectID(5), uint32(48), "test123").
		Return(&index.Subject{}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := test.New(t).
		Put("/v0/indices/7/subjects/5").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]any{
			"sort":    48,
			"comment": "test123",
		}).
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_Update_Index_Subject_NoPermission(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{
		CreatorID: 1,
		ID:        7,
	}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := test.New(t).
		Put("/v0/indices/7/subjects/5").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]any{
			"sort":    48,
			"comment": "test123",
		}).
		Execute(app)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_Update_Index_Subject_NonExists(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 1}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{
		CreatorID: 1,
		ID:        7,
	}, nil)
	mockIndex.EXPECT().AddOrUpdateIndexSubject(mock.Anything, uint32(7), model.SubjectID(5), uint32(48), "test123").
		Return(&index.Subject{}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := test.New(t).
		Put("/v0/indices/7/subjects/5").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]any{
			"sort":    48,
			"comment": "test123",
		}).
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_Delete_Index_Subject(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{
		CreatorID: 6,
		ID:        7,
	}, nil)
	mockIndex.EXPECT().
		DeleteIndexSubject(mock.Anything, model.IndexID(7), model.SubjectID(5)).
		Return(nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := test.New(t).
		Delete("/v0/indices/7/subjects/5").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_Delete_Index_Subject_NoPermission(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{
		CreatorID: 1,
		ID:        7,
	}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := test.New(t).
		Delete("/v0/indices/7/subjects/5").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_Update_Index_Invalid_Comment(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	app := test.GetWebApp(t, test.Mock{AuthRepo: mockAuth})

	resp := test.New(t).
		Put("/v0/indices/7/subjects/5").
		Header(fiber.HeaderAuthorization, "Bearer token").
		JSON(map[string]any{
			"sort":    48,
			"comment": "test123\000",
		}).
		Execute(app)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
