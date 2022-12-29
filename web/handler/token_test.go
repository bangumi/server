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

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/web/session"
)

func TestHandler_CreatePersonalAccessToken(t *testing.T) {
	t.Parallel()
	const userID model.UserID = 1

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().CreateAccessToken(mock.Anything, userID, "token name", gtime.OneDay).Return("ttt", nil)
	mockAuth.EXPECT().GetByID(mock.Anything, mock.Anything).Return(auth.Auth{ID: userID}, nil)

	mockSession := mocks.NewSessionManager(t)
	mockSession.EXPECT().Get(mock.Anything, "session key").Return(session.Session{UserID: userID}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthService:    mockAuth,
			SessionManager: mockSession,
		},
	)

	resp := test.New(t).Post("/p/access-tokens").JSON(fiber.Map{
		"name":          "token name",
		"duration_days": 1,
	}).Cookie(session.CookieKey, "session key").Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode, resp.BodyString())
}

func TestHandler_DeletePersonalAccessToken_401(t *testing.T) {
	t.Parallel()
	const userID model.UserID = 1
	const tokenID uint32 = 5

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetTokenByID(mock.Anything, tokenID).Return(auth.AccessToken{UserID: 2, ID: tokenID}, nil)
	mockAuth.EXPECT().GetByID(mock.Anything, mock.Anything).Return(auth.Auth{ID: userID}, nil)

	mockSession := mocks.NewSessionManager(t)
	mockSession.EXPECT().Get(mock.Anything, "session key").Return(session.Session{UserID: userID}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthService:    mockAuth,
			SessionManager: mockSession,
		},
	)

	resp := test.New(t).Delete("/p/access-tokens").JSON(fiber.Map{"id": tokenID}).
		Cookie(session.CookieKey, "session key").Execute(app)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, resp.BodyString())
}
