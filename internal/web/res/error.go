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
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	"github.com/bangumi/server/internal/web/util"
)

var ErrNotFound = NewError(http.StatusNotFound, "resource can't be found in the database or has been removed")

// Error default error response.
type Error struct {
	Title       string      `json:"title"`
	Details     interface{} `json:"details,omitempty"`
	Description string      `json:"description"`
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

func FromError(c *fiber.Ctx, err error, code int, message string) error {
	return JSON(c.Status(code), Error{
		Title:       utils.StatusMessage(code),
		Description: message,
		Details:     util.DetailWithErr(c, err),
	})
}

func InternalError(c *fiber.Ctx, err error, message string) error {
	return JSON(c.Status(http.StatusInternalServerError), Error{
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
