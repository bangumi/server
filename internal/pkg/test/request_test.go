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

package test_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/test"
)

type res struct {
	Q string `json:"q"`
	I int    `json:"i"`
}

func TestClientFullExample(t *testing.T) {
	t.Parallel()
	app := echo.New()

	app.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, res{I: 5, Q: c.QueryParam("q")})
	})

	var r res
	test.New(t).Get("/test").Query("q", "v").
		Execute(app).
		JSON(&r).
		ExpectCode(http.StatusOK)

	require.Equal(t, 5, r.I)
	require.Equal(t, "v", r.Q)
}
