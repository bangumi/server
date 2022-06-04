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

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/res/code"
)

func (h Handler) translationValidationError(err error) []string {
	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); ok {
		var details = make([]string, len(validationErrors))
		for i, e := range validationErrors {
			// can translate each error one at a time.
			details[i] = e.Translate(h.validatorTranslation)
		}

		return details
	}

	return []string{err.Error()}
}

func (h Handler) InternalServerError(c *fiber.Ctx, err error, message string) error {
	h.skip1Log.Error("internal server error", zap.Error(err))
	return res.JSON(c.Status(code.InternalServerError), res.Error{
		Title:       "Internal Server Error",
		Description: message,
		Details:     err.Error(),
	})
}
