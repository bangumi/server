// Copyright (c) 2021-2022 Trim21 <trim21.me@gmail.com>
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

//go:build !dev

package recovery

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/logger"
)

var errInternal = errors.New("internal server error")

// New creates a new middleware handler.
func New() fiber.Handler {
	// Set default config
	log := logger.Named("http.recovery")
	// Return new handler
	return func(c *fiber.Ctx) (err error) { //nolint:nonamedreturns
		defer func() {
			if r := recover(); r != nil {
				log.Error("recovery", zap.Any("recovery", r))

				var ok bool
				if err, ok = r.(error); !ok {
					// Set error that will call the global error handler
					err = fmt.Errorf("%w: %v", errInternal, r)
				}
			}
		}()

		// Return errInternal if exist, else move to next handler
		return c.Next()
	}
}
