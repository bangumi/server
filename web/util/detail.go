// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	"github.com/bangumi/server/web/res"
)

func DetailFromRequest(c *fiber.Ctx) Detail {
	return Detail{
		Path:        c.Path(),
		QueryString: utils.UnsafeString(c.Request().URI().QueryString()),
	}
}

type Detail struct {
	Error       string `json:"error,omitempty"`
	Path        string `json:"path,omitempty"`
	QueryString string `json:"query_string,omitempty"`
}

func Error(c *fiber.Ctx, code int, message string) error {
	return c.Status(http.StatusBadRequest).JSON(res.Error{
		Title:       utils.StatusMessage(code),
		Description: message,
	})
}
