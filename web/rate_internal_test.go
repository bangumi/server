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

package web

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/rate"
	"github.com/bangumi/server/web/rate/action"
)

func Test_rateMiddleware(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	r := mocks.NewRateLimiter(t)
	r.EXPECT().AllowAction(mock.Anything, model.UserID(1), mock.Anything, mock.Anything).
		Return(false, 1, nil)

	app.Use(rateMiddleware(r, mockBaseHandler{
		a: &accessor.Accessor{
			RequestID: "fake-request-id", IP: net.IPv4(1, 1, 1, 1), Auth: auth.Auth{ID: 1}, Login: true,
		},
	}, action.Action(0), rate.PerHour(10)))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	res, err := app.Test(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusTooManyRequests, res.StatusCode)
}

func Test_rateMiddleware_allow(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	r := mocks.NewRateLimiter(t)
	r.EXPECT().AllowAction(mock.Anything, model.UserID(1), mock.Anything, mock.Anything).
		Return(true, 1, nil)

	app.Use(rateMiddleware(r, mockBaseHandler{
		a: &accessor.Accessor{
			RequestID: "fake-request-id", IP: net.IPv4(1, 1, 1, 1), Auth: auth.Auth{ID: 1}, Login: true,
		},
	}, action.Action(0), rate.PerHour(10)))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("")
	})

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	res, err := app.Test(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func Test_rateMiddleware_not_login(t *testing.T) {
	t.Parallel()
	app := fiber.New(fiber.Config{ErrorHandler: getDefaultErrorHandler()})

	r := mocks.NewRateLimiter(t)

	app.Use(rateMiddleware(r, mockBaseHandler{
		a: &accessor.Accessor{
			RequestID: "fake-request-id", IP: net.IPv4(1, 1, 1, 1), Auth: auth.Auth{ID: 1}, Login: false,
		},
	}, action.Action(0), rate.PerHour(10)))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	res, err := app.Test(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

type mockBaseHandler struct {
	a *accessor.Accessor
}

func (h mockBaseHandler) GetHTTPAccessor(c *fiber.Ctx) *accessor.Accessor {
	return h.a
}
