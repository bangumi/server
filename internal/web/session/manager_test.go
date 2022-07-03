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
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/web/session"
)

func TestManager_Create(t *testing.T) {
	t.Parallel()
	const uid model.UserID = 1

	m := session.NewMockRepo(t)
	m.EXPECT().Create(mock.Anything, mock.Anything, uid, mock.Anything).
		Run(func(ctx context.Context, key string, userID model.UserID, regTime time.Time) {
			require.Equal(t, uid, userID)
		}).Return(session.Session{UserID: uid}, nil)

	manager := session.New(test.NopCache(), m, logger.Copy())

	_, s, err := manager.Create(context.Background(), domain.Auth{ID: uid})
	require.NoError(t, err)
	require.Equal(t, uid, s.UserID)
}

func TestManager_Get(t *testing.T) {
	t.Parallel()
	const uid model.UserID = 1
	m := session.NewMockRepo(t)
	m.EXPECT().Create(mock.Anything, mock.Anything, uid, mock.Anything).
		Run(func(ctx context.Context, key string, userID model.UserID, regTime time.Time) {
			require.Equal(t, uid, userID)
		}).Return(session.Session{UserID: uid}, nil)

	manager := session.New(test.NopCache(), m, logger.Copy())

	_, s, err := manager.Create(context.Background(), domain.Auth{ID: uid})
	require.NoError(t, err)
	require.Equal(t, uid, s.UserID)
}

func TestManager_Revoke(t *testing.T) {
	t.Parallel()

	const uid model.UserID = 1
	m := session.NewMockRepo(t)
	m.EXPECT().Create(mock.Anything, mock.Anything, uid, mock.Anything).
		Run(func(ctx context.Context, key string, userID model.UserID, regTime time.Time) {
			require.Equal(t, uid, userID)
		}).Return(session.Session{UserID: uid}, nil)

	manager := session.New(test.NopCache(), m, logger.Copy())

	_, s, err := manager.Create(context.Background(), domain.Auth{ID: uid})
	require.NoError(t, err)
	require.Equal(t, uid, s.UserID)
}
