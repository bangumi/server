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
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gookit/goutil/timex"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/internal/web/session"
)

func TestHandler_CreatePersonalAccessToken(t *testing.T) {
	t.Parallel()
	const userID model.UIDType = 1

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().CreateAccessToken(mock.Anything, userID, "token name", timex.OneDay).Return("ttt", nil)
	mockAuth.EXPECT().GetByIDWithCache(mock.Anything, mock.Anything).Return(domain.Auth{ID: userID}, nil)

	mockSession := mocks.NewSessionManager(t)
	mockSession.EXPECT().Get(mock.Anything, "session key").Return(session.Session{UserID: userID}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthService:    mockAuth,
			SessionManager: mockSession,
		},
	)

	resp := test.New(t).Post("/p/access-tokens").Header(fiber.HeaderOrigin, config.FrontendOrigin).JSON(fiber.Map{
		"name":          "token name",
		"duration_days": 1,
	}).Cookie(session.Key, "session key").Execute(app, -1)

	require.Equal(t, fiber.StatusOK, resp.StatusCode, resp.BodyString())
}

func TestHandler_DeletePersonalAccessToken_401(t *testing.T) {
	t.Parallel()
	const userID model.UIDType = 1
	const tokenID uint32 = 5

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetTokenByID(mock.Anything, tokenID).Return(domain.AccessToken{UserID: 2, ID: tokenID}, nil)
	mockAuth.EXPECT().GetByIDWithCache(mock.Anything, mock.Anything).Return(domain.Auth{ID: userID}, nil)

	mockSession := mocks.NewSessionManager(t)
	mockSession.EXPECT().Get(mock.Anything, "session key").Return(session.Session{UserID: userID}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthService:    mockAuth,
			SessionManager: mockSession,
		},
	)

	resp := test.New(t).Delete("/p/access-tokens").Header(fiber.HeaderOrigin, config.FrontendOrigin).JSON(fiber.Map{
		"id": tokenID,
	}).Cookie(session.Key, "session key").Execute(app, -1)

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, resp.BodyString())
}
