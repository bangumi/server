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

package session_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/web/session"
)

func TestManager_Create(t *testing.T) {
	t.Parallel()
	const uid model.UserID = 1

	m := mocks.NewSessionRepo(t)
	m.EXPECT().Create(mock.Anything, uid, mock.Anything, mock.Anything).
		Return("", session.Session{UserID: uid}, nil)

	manager := session.New(cache.NewNoop(), m, logger.Copy())

	_, s, err := manager.Create(context.Background(), auth.Auth{ID: uid})
	require.NoError(t, err)
	require.Equal(t, uid, s.UserID)
}

func TestManager_Get(t *testing.T) {
	t.Parallel()
	const uid model.UserID = 1
	m := mocks.NewSessionRepo(t)
	m.EXPECT().Create(mock.Anything, uid, mock.Anything, mock.Anything).
		Return("", session.Session{UserID: uid}, nil)

	manager := session.New(cache.NewNoop(), m, logger.Copy())

	_, s, err := manager.Create(context.Background(), auth.Auth{ID: uid})
	require.NoError(t, err)
	require.Equal(t, uid, s.UserID)
}

func TestManager_Revoke(t *testing.T) {
	t.Parallel()

	const uid model.UserID = 1
	m := mocks.NewSessionRepo(t)
	m.EXPECT().Create(mock.Anything, uid, mock.Anything, mock.Anything).
		Return("", session.Session{UserID: uid}, nil)

	manager := session.New(cache.NewNoop(), m, logger.Copy())

	_, s, err := manager.Create(context.Background(), auth.Auth{ID: uid})
	require.NoError(t, err)
	require.Equal(t, uid, s.UserID)
}
