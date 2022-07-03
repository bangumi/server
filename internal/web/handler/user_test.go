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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/web/res"
)

func TestHandler_GetCurrentUser(t *testing.T) {
	t.Parallel()
	const uid model.UserID = 7

	u := mocks.NewUserRepo(t)
	u.EXPECT().GetByID(mock.Anything, uid).Return(model.User{ID: uid}, nil)

	a := mocks.NewAuthRepo(t)
	a.EXPECT().GetByToken(mock.Anything, "token").Return(domain.Auth{ID: uid}, nil)
	a.EXPECT().GetPermission(mock.Anything, mock.Anything).Return(domain.Permission{}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo: a,
			UserRepo: u,
		},
	)

	var r res.User
	resp := test.New(t).Get("/v0/me").Header("authorization", "Bearer token").
		Execute(app).JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.EqualValues(t, uid, r.ID, resp.BodyString())
}

func TestHandler_GetUser_200(t *testing.T) {
	t.Parallel()
	const uid model.UserID = 7
	m := mocks.NewUserRepo(t)
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

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, mock.Anything).Return(model.User{}, domain.ErrNotFound)

	app := test.GetWebApp(t,
		test.Mock{
			UserRepo: m,
		},
	)

	resp := test.New(t).Get("/v0/users/u").Execute(app)
	require.Equal(t, http.StatusNotFound, resp.StatusCode, resp.BodyString())
}

func TestHandler_GetUserAvatar_302(t *testing.T) {
	t.Parallel()

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, "u").Return(model.User{ID: 1, Avatar: "temp"}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: m})
	for _, imageType := range []string{"large", "medium", "small"} {
		t.Run(imageType, func(t *testing.T) {
			t.Parallel()

			resp := test.New(t).Get("/v0/users/u/avatar?type=" + imageType).Execute(app)
			require.Equal(t, http.StatusFound, resp.StatusCode, resp.BodyString())
			expected, _ := res.UserAvatar("temp").Select(imageType)
			require.Equal(t, expected, resp.Header.Get("Location"), "expect redirect to image url")
		})
	}
}

func TestHandler_GetUserAvatar_400(t *testing.T) {
	t.Parallel()

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, mock.Anything).Return(model.User{Avatar: "temp"}, nil)
	app := test.GetWebApp(t,
		test.Mock{
			UserRepo: m,
		},
	)

	resp := test.New(t).Get("/v0/users/u/avatar").Execute(app)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, resp.BodyString())
}
