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
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/logger"
)

func TestOnUserChangeRevokesTokensOnACLChange(t *testing.T) {
	t.Parallel()

	const userID = model.UserID(1)

	for _, testCase := range []struct {
		name   string
		before string
		after  string
	}{
		{
			name:   "acl",
			before: `{"acl":"a:0:{}","groupid":1}`,
			after:  `{"acl":"a:1:{s:8:\"ban_post\";s:1:\"1\";}","groupid":1}`,
		},
		{
			name:   "group",
			before: `{"acl":"a:0:{}","groupid":1}`,
			after:  `{"acl":"a:0:{}","groupid":2}`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			a := mocks.NewAuthService(t)
			a.EXPECT().RevokeUserTokens(mock.Anything, userID).Return(nil)

			e := eventHandler{auth: a, log: logger.Copy()}
			err := e.OnUserChange(context.Background(), json.RawMessage(`{"uid":1}`), Payload{
				Op:     opUpdate,
				Before: json.RawMessage(testCase.before),
				After:  json.RawMessage(testCase.after),
			})
			require.NoError(t, err)
		})
	}
}

func TestOnUserChangeIgnoresUnrelatedUpdate(t *testing.T) {
	t.Parallel()

	e := eventHandler{auth: mocks.NewAuthService(t), log: logger.Copy()}
	err := e.OnUserChange(context.Background(), json.RawMessage(`{"uid":1}`), Payload{
		Op:     opUpdate,
		Before: json.RawMessage(`{"acl":"a:0:{}","groupid":1,"avatar":"a.png"}`),
		After:  json.RawMessage(`{"acl":"a:0:{}","groupid":1,"avatar":"a.png"}`),
	})
	require.NoError(t, err)
}

func TestOnUserChangeReturnsTokenRevokeError(t *testing.T) {
	t.Parallel()

	a := mocks.NewAuthService(t)
	a.EXPECT().RevokeUserTokens(mock.Anything, model.UserID(1)).
		Return(errors.New("redis unavailable"))

	e := eventHandler{auth: a, log: logger.Copy()}
	err := e.OnUserChange(context.Background(), json.RawMessage(`{"uid":1}`), Payload{
		Op:     opUpdate,
		Before: json.RawMessage(`{"acl":"a:0:{}","groupid":1}`),
		After:  json.RawMessage(`{"acl":"a:1:{s:8:\"ban_post\";s:1:\"1\";}","groupid":1}`),
	})
	require.Error(t, err)
}
