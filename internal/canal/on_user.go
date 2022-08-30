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
	"github.com/goccy/go-json"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func (e *eventHandler) OnUserChange(key json.RawMessage, payload payload) error {
	switch payload.Op {
	case opCreate, opSnapshot, opDelete:
		return nil
	case opUpdate:
		var before userPayload
		if err := json.Unmarshal(payload.Before, &before); err != nil {
			return errgo.Wrap(err, "json")
		}
		var after userPayload
		if err := json.Unmarshal(payload.After, &after); err != nil {
			return errgo.Wrap(err, "json")
		}

		if before.Password != after.Password {
			var k UserKey
			if err := json.UnmarshalNoEscape(key, &k); err != nil {
				e.log.Error("failed to unmarshal json", zap.Error(err))
				return errgo.Wrap(err, "json.Unmarshal")
			}
			return e.OnUserPasswordChange(k.ID)
		}
	}

	return nil
}

type UserKey struct {
	ID model.UserID `json:"uid"`
}

type userPayload struct {
	Password string `json:"password_crypt"`
}
