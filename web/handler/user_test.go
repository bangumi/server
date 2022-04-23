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

package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/mocks"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/res"
)

func TestHandler_GetCurrentUser(t *testing.T) {
	t.Parallel()
	const uid model.IDType = 7
	u := mocks.UserRepo{}
	u.EXPECT().GetByID(mock.Anything, uid).Return(model.User{ID: uid}, nil)
	defer u.AssertExpectations(t)

	a := mocks.AuthRepo{}
	a.EXPECT().GetByToken(mock.Anything, "token").Return(domain.Auth{ID: uid}, nil)
	defer a.AssertExpectations(t)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo: &a,
			UserRepo: &u,
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/v0/me", http.NoBody)
	req.Header.Set("authorization", "Bearer token")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var r res.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&r))
	require.Equal(t, uid, r.ID)
}

func TestHandler_GetUser(t *testing.T) {
	t.Parallel()
	const uid model.IDType = 7
	m := &mocks.UserRepo{}
	m.EXPECT().GetByName(mock.Anything, "u").Return(model.User{ID: uid}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			UserRepo: m,
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/v0/users/u", http.NoBody)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var r res.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&r))
	require.Equal(t, uid, r.ID)
}
