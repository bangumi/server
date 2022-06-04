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
