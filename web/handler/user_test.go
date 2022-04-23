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
	"testing"

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

	var r res.User
	resp := test.New(t).Get("/v0/me").Header("authorization", "Bearer token").
		Execute(app).JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, uid, r.ID)
}

func TestHandler_GetUser_200(t *testing.T) {
	t.Parallel()
	const uid model.IDType = 7
	m := &mocks.UserRepo{}
	m.EXPECT().GetByName(mock.Anything, "u").Return(model.User{ID: uid}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			UserRepo: m,
		},
	)

	var r res.User
	resp := test.New(t).Get("/v0/users/u").Execute(app).JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, uid, r.ID)
}

func TestHandler_GetUser_404(t *testing.T) {
	t.Parallel()

	m := &mocks.UserRepo{}
	m.EXPECT().GetByName(mock.Anything, mock.Anything).Return(model.User{}, domain.ErrNotFound)
	defer m.AssertExpectations(t)

	app := test.GetWebApp(t,
		test.Mock{
			UserRepo: m,
		},
	)

	resp := test.New(t).Get("/v0/users/u").Execute(app)
	require.Equal(t, http.StatusNotFound, resp.StatusCode, resp.BodyString())
}
