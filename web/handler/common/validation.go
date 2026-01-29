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

package common

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/labstack/echo/v5"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/web/res"
)

func (h Common) ValidationError(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, res.Error{
		Title:       "Bad Request",
		Description: "can't validate request body",
		Details:     h.translationValidationError(err),
	})
}

func (h Common) translationValidationError(err error) []string {
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

var errTranslationNotFound = errors.New("failed to find translation for zh")

func getValidator() (*validator.Validate, ut.Translator, error) {
	validate := validator.New()
	uni := ut.New(en.New(), zh.New())

	// this is usually know or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	trans, found := uni.GetTranslator(zh.New().Locale())
	if !found {
		return nil, nil, errTranslationNotFound
	}

	err := zhTranslations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		return nil, nil, errgo.Wrap(err, "failed to register translation")
	}

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		docName := field.Tag.Get("validateName")
		if docName != "" {
			return docName
		}

		tag := field.Tag.Get("json")
		if tag == "" {
			return field.Name
		}

		name := strings.SplitN(tag, ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})

	return validate, trans, nil
}
