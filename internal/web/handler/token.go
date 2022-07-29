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
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/pkg/timex"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) CreatePersonalAccessToken(c *fiber.Ctx) error {
	v := h.GetHTTPAccessor(c)
	if !v.Login {
		return c.Redirect("/demo/login")
	}

	var r req.CreatePersonalAccessToken
	if err := json.UnmarshalNoEscape(c.Body(), &r); err != nil {
		return res.JSONError(c, err)
	}

	if err := h.Common.V.Struct(r); err != nil {
		return h.ValidationError(c, err)
	}

	token, err := h.a.CreateAccessToken(c.Context(), v.ID, r.Name, timex.OneDay*time.Duration(r.DurationDays))
	if err != nil {
		return h.InternalError(c, err, "failed to create token", log.UserID(v.ID), zap.String("token_name", r.Name))
	}

	return c.JSON(token)
}

func (h Handler) DeletePersonalAccessToken(c *fiber.Ctx) error {
	v := h.GetHTTPAccessor(c)
	if !v.Login {
		return c.Redirect("/demo/login")
	}

	var r req.DeletePersonalAccessToken
	if err := json.UnmarshalNoEscape(c.Body(), &r); err != nil {
		return res.JSONError(c, err)
	}
	if err := h.Common.V.Struct(r); err != nil {
		return h.ValidationError(c, err)
	}

	token, err := h.a.GetTokenByID(c.Context(), r.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.BadRequest("token not exist")
		}

		return h.InternalError(c, err, "failed to get token info", zap.Uint32("token_id", r.ID))
	}

	if token.UserID != v.ID {
		return res.Unauthorized("you don't have this token")
	}

	ok, err := h.a.DeleteAccessToken(c.Context(), r.ID)
	if err != nil {
		return h.InternalError(c, err, "failed to delete token", zap.Uint32("token_id", r.ID), log.UserID(v.ID))
	}

	if !ok {
		return c.SendStatus(http.StatusNotFound)
	}

	return c.SendStatus(http.StatusNoContent)
}
