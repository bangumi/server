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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/web/res"
)

func TestDefaultErrorHandler_resError(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{ErrorHandler: getDefaultErrorHandler()})
	app.Get("/", func(c *fiber.Ctx) error {
		return res.BadRequest("mm")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body res.Error
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	require.Equal(t, "mm", body.Description)
}

func TestDefaultErrorHandler_internal(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{ErrorHandler: getDefaultErrorHandler()})
	app.Get("/", func(c *fiber.Ctx) error {
		return errors.New("mm") //nolint:goerr113
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
