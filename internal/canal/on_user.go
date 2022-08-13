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
	"github.com/bangumi/server/internal/pkg/generic/slice"
)

func (e *eventHandler) OnUserChange(key json.RawMessage, payload payload) {
	switch payload.Op {
	case opCreate, opReplace, opDelete:
		return
	case opUpdate:
		var diff = make([]string, 0, len(payload.After))
		for k, v := range payload.Before {
			if string(payload.After[k]) != string(v) {
				diff = append(diff, k)
			}
		}

		if slice.Contains(diff, "password_crypt") {
			var k UserKey
			if err := json.UnmarshalNoEscape(key, &k); err != nil {
				e.log.Error("failed to unmarshal json", zap.Error(err))
				return
			}
			e.OnUserPasswordChange(k.ID)
		}
	}
}

type UserKey struct {
	ID model.UserID `json:"uid"`
}
