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

package notification_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/web/session"
)

func TestNotification_Count(t *testing.T) {
	t.Parallel()
	m := mocks.NewNotificationRepo(t)
	m.EXPECT().Count(
		mock.Anything,
		model.UserID(1),
	).Return(0, nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByID(mock.Anything, mock.Anything).Return(auth.Auth{ID: 1}, nil)

	s := mocks.NewSessionManager(t)
	s.EXPECT().Get(mock.Anything, "11").Return(session.Session{UserID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{NotificationRepo: m, AuthService: mockAuth, SessionManager: s})

	resp := htest.New(t, app).
		Header(echo.HeaderCookie, "chiiNextSessionID=11").
		Get("/p/notifications/count").
		ExpectCode(http.StatusOK)

	var v int
	err := json.Unmarshal(resp.Body, &v)
	require.NoError(t, err)
}
