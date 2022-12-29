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
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/handler/internal/ctxkey"
	"github.com/bangumi/server/web/session"
)

func New(
	log *zap.Logger,
	auth auth.Service,
	session session.Manager,
	config config.AppConfig,
) (Common, error) {
	validate, trans, err := getValidator()
	if err != nil {
		return Common{}, err
	}

	log = log.Named("handler.Common")
	return Common{
		Config:               config,
		session:              session,
		auth:                 auth,
		log:                  log,
		skip1Log:             log.WithOptions(zap.AddCallerSkip(1)),
		V:                    validate,
		validatorTranslation: trans,
	}, nil
}

type Common struct {
	Config               config.AppConfig
	auth                 auth.Service
	skip1Log             *zap.Logger
	log                  *zap.Logger
	session              session.Manager
	V                    *validator.Validate
	validatorTranslation ut.Translator
}

func (h Common) GetHTTPAccessor(c *fiber.Ctx) *accessor.Accessor {
	u, ok := c.Context().UserValue(ctxkey.User).(*accessor.Accessor)
	if !ok {
		h.log.Error(
			"failed to get http accessor, expecting *accessor got another type instead",
			zap.Any("accessor", c.Context().UserValue(ctxkey.User)))
		panic("can't convert type")
	}

	return u
}
