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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/web/res/code"
	"github.com/bangumi/server/internal/web/util"
)

// Error default error response.
type Error struct {
	Title       string      `json:"title"`
	Details     interface{} `json:"details,omitempty"`
	Description string      `json:"description"`
}

func WithError(c *fiber.Ctx, err error, code int, message string) error {
	return JSON(c.Status(code), Error{
		Title:       http.StatusText(code),
		Description: message,
		Details:     util.ErrDetail(c, err),
	})
}

func HTTPError(c *fiber.Ctx, code int, message string) error {
	return JSON(c.Status(code), Error{
		Title:       http.StatusText(code),
		Description: message,
	})
}

func InternalError(c *fiber.Ctx, err error, message string) error {
	return JSON(c.Status(code.InternalServerError), Error{
		Title:       "Internal Server Error",
		Description: message,
		Details:     err.Error(),
	})
}

const DefaultUnauthorizedMessage = "you are not allowed to do this"

func Unauthorized(c *fiber.Ctx, message string) error {
	return JSON(c.Status(code.Unauthorized), Error{
		Title:       "Unauthorized",
		Description: message,
	})
}

func NotFound(c *fiber.Ctx, message string) error {
	return JSON(c.Status(code.NotFound), Error{
		Title:       "Not Found",
		Description: message,
		Details:     util.DetailFromRequest(c),
	})
}
