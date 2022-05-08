// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
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

package session_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/mocks"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/session"
)

func TestManager_Create(t *testing.T) {
	t.Parallel()
	const uid model.UIDType = 1

	m := &mocks.SessionRepo{}
	m.EXPECT().Create(mock.Anything, mock.Anything, uid, mock.AnythingOfType("session.Session")).
		Run(func(ctx context.Context, key string, userID uint32, s session.Session) {
			require.Equal(t, uid, s.UserID)
		}).Return(nil)
	defer m.AssertExpectations(t)

	manager := session.New(test.NopCache(), m, logger.Copy())

	_, s, err := manager.Create(context.Background(), domain.Auth{ID: uid})
	require.NoError(t, err)
	require.Equal(t, uid, s.UserID)
}

func TestManager_Get(t *testing.T) {
	t.Parallel()
	const uid model.UIDType = 1
	m := &mocks.SessionRepo{}
	m.EXPECT().Create(mock.Anything, mock.Anything, uid, mock.AnythingOfType("session.Session")).
		Run(func(ctx context.Context, key string, userID uint32, s session.Session) {
			require.Equal(t, uid, s.UserID)
		}).Return(nil)
	defer m.AssertExpectations(t)

	manager := session.New(test.NopCache(), m, logger.Copy())

	_, s, err := manager.Create(context.Background(), domain.Auth{ID: uid})
	require.NoError(t, err)
	require.Equal(t, uid, s.UserID)

}

func TestManager_Revoke(t *testing.T) {
	t.Parallel()

	const uid model.UIDType = 1
	m := &mocks.SessionRepo{}
	m.EXPECT().Create(mock.Anything, mock.Anything, uid, mock.AnythingOfType("session.Session")).
		Run(func(ctx context.Context, key string, userID uint32, s session.Session) {
			require.Equal(t, uid, s.UserID)
		}).Return(nil)
	defer m.AssertExpectations(t)

	manager := session.New(test.NopCache(), m, logger.Copy())

	_, s, err := manager.Create(context.Background(), domain.Auth{ID: uid})
	require.NoError(t, err)
	require.Equal(t, uid, s.UserID)
}
