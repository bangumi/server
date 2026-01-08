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

package index_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestCollectIndex(t *testing.T) {
	t.Parallel()
	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(233)).Return(model.Index{ID: 233}, nil)
	mockIndex.EXPECT().GetIndexCollect(mock.Anything, mock.Anything, mock.Anything).Return(nil, gerr.ErrNotFound)
	mockIndex.EXPECT().AddIndexCollect(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).Return(auth.Permission{}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		Post("/v0/indices/233/collect")

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUncollectIndex(t *testing.T) {
	t.Parallel()
	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(322)).Return(model.Index{ID: 322}, nil)
	mockIndex.EXPECT().GetIndexCollect(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mockIndex.EXPECT().DeleteIndexCollect(mock.Anything, uint32(322), uint32(6)).Return(nil)
	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).Return(auth.Permission{}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		Delete("/v0/indices/322/collect")

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCollectIndex_PrivateNotOwner(t *testing.T) {
	t.Parallel()
	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(233)).Return(
		model.Index{ID: 233, CreatorID: 1, Privacy: model.IndexPrivacyPrivate},
		nil,
	)
	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).Return(auth.Permission{}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		Post("/v0/indices/233/collect")

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
