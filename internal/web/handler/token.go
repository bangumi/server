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

package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gookit/goutil/timex"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/res/code"
)

func (h Handler) CreatePersonalAccessToken(c *fiber.Ctx) error {
	v := h.getHTTPAccessor(c)
	if !v.login {
		return c.Redirect("/demo/login")
	}

	var r req.CreatePersonalAccessToken
	if err := json.UnmarshalNoEscape(c.Body(), &r); err != nil {
		return res.WithError(c, err, code.UnprocessableEntity, "can't parse request body as JSON")
	}

	if err := h.v.Struct(r); err != nil {
		return res.JSON(c.Status(code.BadRequest), res.Error{
			Title:       http.StatusText(code.BadRequest),
			Description: "can't validate request body",
			Details:     h.translationValidationError(err),
		})
	}

	token, err := h.a.CreateAccessToken(c.Context(), v.ID, r.Name, timex.OneDay*time.Duration(r.DurationDays))
	if err != nil {
		return res.InternalError(c, err, "failed to create token")
	}

	return c.JSON(token)
}

func (h Handler) DeletePersonalAccessToken(c *fiber.Ctx) error {
	v := h.getHTTPAccessor(c)
	if !v.login {
		return c.Redirect("/demo/login")
	}

	var r req.DeletePersonalAccessToken
	if err := json.UnmarshalNoEscape(c.Body(), &r); err != nil {
		return res.WithError(c, err, code.UnprocessableEntity, "can't parse request body as JSON")
	}

	if err := h.v.Struct(r); err != nil {
		return res.JSON(c.Status(code.BadRequest), res.Error{
			Title:       http.StatusText(code.BadRequest),
			Description: "can't validate request body",
			Details:     h.translationValidationError(err),
		})
	}

	token, err := h.a.GetTokenByID(c.Context(), r.ID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return res.InternalError(c, err, "failed to get token info")
	}

	if token.UserID != v.ID {
		return res.HTTPError(c, code.Unauthorized, "you don't have this token")
	}

	ok, err := h.a.DeleteAccessToken(c.Context(), r.ID)
	if err != nil {
		return res.InternalError(c, err, "failed to create token")
	}

	if !ok {
		return c.SendStatus(code.NotFound)
	}

	return c.SendStatus(http.StatusNoContent)
}
