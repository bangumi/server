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

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/web/middleware/recovery"
)

func TestPanicMiddleware(t *testing.T) {
	t.Parallel()
	var app = fiber.New()

	app.Use(recovery.New())

	app.Get("/", func(c *fiber.Ctx) error {
		panic("errInternal")
	})

	req := httptest.NewRequest("GET", "/", nil)

	resp, err := app.Test(req)
	require.Nil(t, err, "panic should be caught")
	defer resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError,
		resp.StatusCode, "middleware should catch internal error")
}
