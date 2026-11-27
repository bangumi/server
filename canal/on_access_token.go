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

package canal

import (
	"context"
	"encoding/json"

	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cachekey"
)

type accessTokenPayload struct {
	AccessToken string `json:"access_token"`
}

func (e *eventHandler) OnAccessToken(ctx context.Context, payload Payload) error {
	switch payload.Op {
	case opCreate, opSnapshot:
		return nil
	case opUpdate, opDelete:
		// tokens are revoked by setting expires to now, or deleted after expiration.
		var before accessTokenPayload
		if err := json.Unmarshal(payload.Before, &before); err != nil {
			return errgo.Wrap(err, "json.Unmarshal")
		}

		if before.AccessToken == "" {
			return nil
		}

		if err := e.cache.Del(ctx, cachekey.AuthToken(before.AccessToken)); err != nil {
			return errgo.Wrap(err, "cache.Del")
		}

		return nil
	default:
		e.log.Warn("unexpected operator", zap.String("op", payload.Op))
	}

	return nil
}
