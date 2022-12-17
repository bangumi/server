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
	"fmt"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func (e *eventHandler) OnUserChange(key json.RawMessage, payload payload) error {
	var k UserKey
	if err := sonic.Unmarshal(key, &k); err != nil {
		e.log.Error("failed to unmarshal json", zap.Error(err))
		return errgo.Wrap(err, "sonic.Unmarshal")
	}

	switch payload.Op {
	case opCreate, opSnapshot, opDelete:
		return nil
	case opUpdate:
		var before userPayload
		if err := sonic.Unmarshal(payload.Before, &before); err != nil {
			return errgo.Wrap(err, "json")
		}
		var after userPayload
		if err := sonic.Unmarshal(payload.After, &after); err != nil {
			return errgo.Wrap(err, "json")
		}

		if before.Password != after.Password {
			return e.OnUserPasswordChange(k.ID)
		}

		if before.NewNotify != after.NewNotify {
			e.redis.Publish(context.Background(), fmt.Sprintf("event-user-notify-%d", k.ID), redisUserChannel{
				UserID:    k.ID,
				NewNotify: after.NewNotify,
			})
		}
	}

	return nil
}

type redisUserChannel struct {
	UserID    model.UserID `json:"user_id"`
	NewNotify uint16       `json:"new_notify"`
}

type UserKey struct {
	ID model.UserID `json:"uid"`
}

type userPayload struct {
	Password  string `json:"password_crypt"`
	NewNotify uint16 `json:"new_notify"`
}
