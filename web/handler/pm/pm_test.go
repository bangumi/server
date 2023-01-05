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

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/session"
)

func TestPrivateMessage_List(t *testing.T) {
	t.Parallel()
	m := mocks.NewPrivateMessageRepo(t)
	m.EXPECT().CountByFolder(mock.Anything, model.UserID(1), pm.FolderTypeInbox).Return(1, nil)
	m.EXPECT().List(
		mock.Anything,
		model.UserID(1),
		pm.FolderTypeInbox,
		0,
		10,
	).Return([]pm.PrivateMessageListItem{
		{
			Main: pm.PrivateMessage{},
			Self: pm.PrivateMessage{},
		},
	}, nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByID(mock.Anything, mock.Anything).Return(auth.Auth{ID: 1}, nil)

	s := mocks.NewSessionManager(t)
	s.EXPECT().Get(mock.Anything, "11").Return(session.Session{UserID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth, SessionManager: s})

	resp := test.New(t).
		Get("/p/pms/list?offset=0&limit=10&folder=inbox").
		Header(echo.HeaderCookie, "chiiNextSessionID=11").
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
	).Return([]pm.PrivateMessage{}, domain.ErrNotFound)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByID(mock.Anything, mock.Anything).Return(auth.Auth{ID: 1}, nil)

	s := mocks.NewSessionManager(t)
	s.EXPECT().Get(mock.Anything, "11").Return(session.Session{UserID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth, SessionManager: s})

	resp := test.New(t).
		Get("/p/pms/related-msgs/1").
		Header(echo.HeaderCookie, "chiiNextSessionID=11").
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
	mockAuth.EXPECT().GetByID(mock.Anything, mock.Anything).Return(auth.Auth{ID: 1}, nil)

	s := mocks.NewSessionManager(t)
	s.EXPECT().Get(mock.Anything, "11").Return(session.Session{UserID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth, SessionManager: s})

	resp := test.New(t).
		Get("/p/pms/contacts/recent").
		Header(echo.HeaderCookie, "chiiNextSessionID=11").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPrivateMessage_CountTypes(t *testing.T) {
	t.Parallel()
	m := mocks.NewPrivateMessageRepo(t)
	m.EXPECT().CountTypes(
		mock.Anything,
		model.UserID(1),
	).Return(pm.PrivateMessageTypeCounts{}, nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByID(mock.Anything, mock.Anything).Return(auth.Auth{ID: 1}, nil)

	s := mocks.NewSessionManager(t)
	s.EXPECT().Get(mock.Anything, "111").Return(session.Session{UserID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth, SessionManager: s})

	resp := test.New(t).
		Get("/p/pms/counts").
		Header(echo.HeaderCookie, "chiiNextSessionID=111").
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
	mockAuth.EXPECT().GetByID(mock.Anything, mock.Anything).Return(auth.Auth{ID: 1}, nil)

	s := mocks.NewSessionManager(t)
	s.EXPECT().Get(mock.Anything, "11").Return(session.Session{UserID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth, SessionManager: s})

	resp := test.New(t).
		Patch("/p/pms/read").
		Header(echo.HeaderCookie, "chiiNextSessionID=11").
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
		pm.IDFilter{Type: null.NewFromPtr[model.PrivateMessageID](nil)},
		"测试标题",
		"测试内容",
	).Return([]pm.PrivateMessage{}, nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByID(mock.Anything, mock.Anything).Return(auth.Auth{ID: 1}, nil)

	s := mocks.NewSessionManager(t)
	s.EXPECT().Get(mock.Anything, "111").Return(session.Session{UserID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth, SessionManager: s})

	resp := test.New(t).
		Post("/p/pms").
		Header(echo.HeaderCookie, "chiiNextSessionID=111").
		JSON(req.PrivateMessageCreate{Title: "测试标题", Content: "测试内容", ReceiverIDs: []uint32{382951}}).
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
	mockAuth.EXPECT().GetByID(mock.Anything, mock.Anything).Return(auth.Auth{ID: 1}, nil)

	s := mocks.NewSessionManager(t)
	s.EXPECT().Get(mock.Anything, "111").Return(session.Session{UserID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{PrivateMessageRepo: m, AuthService: mockAuth, SessionManager: s})

	resp := test.New(t).
		Delete("/p/pms").
		Header(echo.HeaderCookie, "chiiNextSessionID=111").
		JSON(req.PrivateMessageDelete{IDs: []uint32{1}}).
		Execute(app)

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}
