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

package util

import (
	"github.com/labstack/echo/v4"
)

func DetailWithErr(c echo.Context, err error) D {
	return D{
		Path:        c.Path(),
		Error:       err.Error(),
		Method:      c.Request().Method,
		QueryString: c.QueryString(),
	}
}

func Detail(c echo.Context) D {
	return D{
		Path:        c.Path(),
		Method:      c.Request().Method,
		QueryString: c.QueryString(),
	}
}

type D struct {
	Error       string `json:"error,omitempty"`
	Path        string `json:"path,omitempty"`
	Method      string `json:"method,omitempty"`
	QueryString string `json:"query_string,omitempty"`
}
