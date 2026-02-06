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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/web/res"
)

func TestDefaultErrorHandler_resError(t *testing.T) {
	t.Parallel()

	app := echo.New()
	app.HTTPErrorHandler = getDefaultErrorHandler()

	app.GET("/", func(c echo.Context) error {
		return res.BadRequest("mm")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)
	require.Equal(t, http.StatusBadRequest, resp.Code)

	content := resp.Body.Bytes()

	var body res.Error
	require.NoError(t, json.Unmarshal(content, &body))

	require.Equal(t, "mm", body.Description)
	require.EqualValues(t, http.StatusBadRequest, resp.Code)
}

func TestDefaultErrorHandler_internal(t *testing.T) {
	t.Parallel()

	app := echo.New()

	app.HTTPErrorHandler = getDefaultErrorHandler()

	app.GET("/", func(c echo.Context) error {
		return errors.New("mm")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	require.Equal(t, http.StatusInternalServerError, resp.Code)
}
