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
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/web/req"
)

func TestPrivateMessage_List(t *testing.T) {
	t.Parallel()
	m := mocks.NewPrivateMessageRepo(t)
	m.EXPECT().List(
		mock.Anything,
		model.UserID(1),
		model.PrivateMessageFolderTypeInbox,
		0,
		10,
	).Return([]model.PrivateMessageListItem{}, nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth})

	resp := test.New(t).
		Get("/v0/pms/list?offset=0&limit=10&folder=inbox").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPrivateMessage_ListRelated(t *testing.T) {
	t.Parallel()
	m := mocks.NewPrivateMessageRepo(t)
	m.EXPECT().ListRelated(
		mock.Anything,
		model.UserID(1),
		model.PrivateMessageID(1),
	).Return([]model.PrivateMessage{}, nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth})

	resp := test.New(t).
		Get("/v0/pms/list/1").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPrivateMessage_ListRecentContact(t *testing.T) {
	t.Parallel()
	m := mocks.NewPrivateMessageRepo(t)
	m.EXPECT().ListRecentContact(
		mock.Anything,
		model.UserID(1),
	).Return([]model.UserID{}, nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth})

	resp := test.New(t).
		Get("/v0/pms/contact/recent").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPrivateMessage_CountTypes(t *testing.T) {
	t.Parallel()
	m := mocks.NewPrivateMessageRepo(t)
	m.EXPECT().CountTypes(
		mock.Anything,
		model.UserID(1),
	).Return(model.PrivateMessageTypeCounts{}, nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth})

	resp := test.New(t).
		Get("/v0/pms/counts").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPrivateMessage_MarkRead(t *testing.T) {
	t.Parallel()
	m := mocks.NewPrivateMessageRepo(t)
	m.EXPECT().MarkRead(
		mock.Anything,
		model.UserID(1),
		model.PrivateMessageID(1),
	).Return(nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth})

	resp := test.New(t).
		Patch("/v0/pms/mark-read").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Header(fiber.HeaderContentType, "application/json").
		JSON(req.PrivateMessageMarkRead{ID: 1}).
		Execute(app)

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestPrivateMessage_Create(t *testing.T) {
	t.Parallel()
	m := mocks.NewPrivateMessageRepo(t)
	m.EXPECT().Create(
		mock.Anything,
		model.UserID(1),
		[]model.UserID{382951},
		domain.PrivateMessageIDFilter{Type: null.NewFromPtr[model.PrivateMessageID](nil)},
		"测试标题",
		"测试内容",
	).Return([]model.PrivateMessage{}, nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth})

	resp := test.New(t).
		Post("/v0/pms/create").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Header(fiber.HeaderContentType, "application/json").
		JSON(req.PrivateMessageCreate{Title: "测试标题", Content: "测试内容", ReceiverIDs: []uint32{382951}, SenderID: 1}).
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPrivateMessage_Delete(t *testing.T) {
	t.Parallel()
	m := mocks.NewPrivateMessageRepo(t)
	m.EXPECT().Delete(
		mock.Anything,
		model.UserID(1),
		[]model.PrivateMessageID{1},
	).Return(nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth})

	resp := test.New(t).
		Delete("/v0/pms/delete").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Header(fiber.HeaderContentType, "application/json").
		JSON(req.PrivateMessageDelete{IDs: []uint32{1}}).
		Execute(app)

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}
