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

package ua

import (
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/web/res"
)

var _ echo.MiddlewareFunc = DisableDefaultHTTPLibrary

const forbiddenMessage = "using HTTP request library's default User-Agent is forbidden, " +
	"please read the document for User-Agent suggestion " +
	"https://github.com/bangumi/api/blob/master/docs-raw/user%20agent.md"

func DisableDefaultHTTPLibrary(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := c.Request().UserAgent()
		if u == "" {
			return res.Forbidden("Please set a 'User-Agent'")
		}

		if isDefaultUA(u) {
			return res.Forbidden(forbiddenMessage)
		}

		return next(c)
	}
}

func isDefaultUA(u string) bool {
	return strings.HasPrefix(u, "Java/") ||
		strings.HasPrefix(u, "axios/") ||
		strings.HasPrefix(u, "okhttp/") ||
		strings.HasPrefix(u, "go-resty/") ||
		strings.HasPrefix(u, "Faraday v") ||
		strings.HasPrefix(u, "node-fetch/") ||
		strings.HasPrefix(u, "Go-http-client/") ||
		strings.HasPrefix(u, "python-requests/") ||
		strings.HasPrefix(u, "Apache-HttpClient/") ||
		u == "node-fetch" ||
		u == "database"
}
