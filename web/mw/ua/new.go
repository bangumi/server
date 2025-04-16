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
	"fmt"
	"regexp"
	"strconv"
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

// DisableBrokenUA disallow known broken app send infinite requests.
func DisableBrokenUA(next echo.HandlerFunc) echo.HandlerFunc {
	aniPattern := regexp.MustCompile(`^open-ani/ani/(\d+.\d+.\d+) .*`)
	return func(c echo.Context) error {
		u := c.Request().UserAgent()
		if u == "" {
			return res.Forbidden("Please set a 'User-Agent'")
		}

		if strings.HasPrefix(u, "open-ani/ani/") {
			m := aniPattern.FindStringSubmatch(u)

			if len(m) < 1 {
				return res.Forbidden(banAnimeko)
			}

			version := m[0]
			fmt.Println(version)
			s := strings.Split(version, ".")

			if len(s) != 3 {
				return res.Forbidden(banAnimeko)
			}

			major, _ := strconv.Atoi(s[0])
			minor, _ := strconv.Atoi(s[1])
			patch, _ := strconv.Atoi(s[2])

			if major <= 4 && minor <= 8 && patch <= 1 {
				return res.Forbidden(banAnimeko)
			}
		}

		return next(c)
	}
}

const banAnimeko = "Animeko version 4.8.1 and earlier contain a bug that causes continuous infinite requests." +
	"Please update to version 4.8.2 or later to resolve this issue."

func isDefaultUA(u string) bool {
	for _, s := range disabledUA {
		if u == s {
			return true
		}
	}

	for _, prefix := range disabledPrefix {
		if strings.HasPrefix(u, prefix) {
			return true
		}
	}

	return false
}

//nolint:gochecknoglobals
var disabledUA = []string{
	"undici",
	"database",
	"node-fetch",
}

//nolint:gochecknoglobals
var disabledPrefix = []string{
	"Java/",
	"axios/",
	"okhttp/",
	"go-resty/",
	"Faraday v",
	"node-fetch/",
	"Go-http-client/",
	"python-requests/",
	"Apache-HttpClient/",
}
