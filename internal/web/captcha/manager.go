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

package captcha

import (
	"context"

	"github.com/go-resty/resty/v2"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/web/captcha/turnstile"
)

type Manager interface {
	Verify(ctx context.Context, response string) (bool, error)
}

func New(cfg config.AppConfig, http *resty.Client) (Manager, error) {
	if cfg.TurnstileSecretKey == "1x0000000000000000000000000000000AA" {
		return nopeManager{}, nil
	}

	return turnstile.New(cfg, http), nil
}
