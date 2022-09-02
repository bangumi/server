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
	"net/http"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestNotification_Count(t *testing.T) {
	t.Parallel()
	m := mocks.NewNotificationRepo(t)
	m.EXPECT().Count(
		mock.Anything,
		model.UserID(1),
	).Return(0, nil)

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{NotificationRepo: m, AuthService: mockAuth})

	resp := test.New(t).
		Get("/v0/notifications/count").
		Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var v int
	err := sonic.Unmarshal(resp.Body, &v)
	require.NoError(t, err)
}
