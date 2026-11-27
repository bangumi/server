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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/pkg/logger"
)

func TestOnAccessTokenChange(t *testing.T) {
	t.Parallel()

	const token = "an-access-token"

	for _, op := range []string{opUpdate, opDelete} {
		t.Run(op, func(t *testing.T) {
			t.Parallel()

			c := mocks.NewRedisCache(t)
			c.EXPECT().Del(mock.Anything, []string{cachekey.AuthToken(token)}).Return(nil)

			e := eventHandler{cache: c, log: logger.Copy()}
			err := e.OnAccessToken(context.Background(), Payload{
				Op:     op,
				Before: json.RawMessage(`{"access_token":"` + token + `"}`),
			})
			require.NoError(t, err)
		})
	}
}

func TestOnAccessTokenCreate(t *testing.T) {
	t.Parallel()

	e := eventHandler{cache: mocks.NewRedisCache(t), log: logger.Copy()}
	err := e.OnAccessToken(context.Background(), Payload{Op: opCreate})
	require.NoError(t, err)
}
