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

package recovery_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/web/mw/recovery"
)

func TestPanicMiddleware(t *testing.T) {
	t.Parallel()
	var app = echo.New()

	app.Use(recovery.New())

	app.GET("/", func(c echo.Context) error {
		panic("errInternal")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	app.ServeHTTP(resp, req)

	require.Equal(t, http.StatusInternalServerError, resp.Code, "middleware should catch internal error")
}
