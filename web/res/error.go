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

package res

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/web/util"
)

var ErrNotFound = NewError(http.StatusNotFound, "resource can't be found in the database or has been removed")

// Error default error response.
type Error struct {
	Title       string `json:"title"`
	Details     any    `json:"details,omitempty"`
	Description string `json:"description"`
}

var _ error = HTTPError{}

type HTTPError struct {
	Msg  string
	Code int
}

func NewError(code int, message string) error {
	return HTTPError{Code: code, Msg: message}
}

func (e HTTPError) Error() string {
	return strconv.Itoa(e.Code) + ": " + e.Msg
}

//nolint:errorlint
func JSONError(c echo.Context, err error) error {
	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		return c.JSON(http.StatusBadRequest, Error{
			Title: "JSON Error",
			Description: fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v",
				ute.Type, ute.Value, ute.Field, ute.Offset),
		})
	}

	if se, ok := err.(*json.SyntaxError); ok {
		return c.JSON(http.StatusBadRequest, Error{
			Title:       "JSON Error",
			Description: fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error()),
		})
	}

	return c.JSON(http.StatusBadRequest, Error{
		Title:       "BodyJSON Error",
		Description: "can't decode request body as json or value doesn't match expected type",
		Details:     util.DetailWithErr(c, err),
	})
}

func InternalError(c echo.Context, err error, message string) error {
	return c.JSON(http.StatusInternalServerError, Error{
		Title:       "Internal Server Error",
		Description: message,
		Details:     util.DetailWithErr(c, err),
	})
}

func BadRequest(message string) error {
	return NewError(http.StatusBadRequest, message)
}

func NotFound(message string) error {
	return NewError(http.StatusNotFound, message)
}

func Unauthorized(message string) error {
	return NewError(http.StatusUnauthorized, message)
}

func Forbidden(message string) error {
	return NewError(http.StatusForbidden, message)
}
