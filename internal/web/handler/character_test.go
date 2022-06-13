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
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/internal/web/res"
)

func TestHandler_GetCharacter_HappyPath(t *testing.T) {
	t.Parallel()
	m := mocks.NewCharacterRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Character{ID: 7}, nil)

	app := test.GetWebApp(t, test.Mock{CharacterRepo: m})

	var r res.CharacterV0
	test.New(t).Get("/v0/characters/7").
		Execute(app).
		JSON(&r).
		ExpectCode(http.StatusOK)
	require.Equal(t, uint32(7), r.ID)
}

func TestHandler_GetCharacter_Redirect(t *testing.T) {
	t.Parallel()
	m := mocks.NewCharacterRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Character{ID: 7, Redirect: 8}, nil)

	app := test.GetWebApp(t, test.Mock{CharacterRepo: m})

	resp := test.New(t).Get("/v0/characters/7").Execute(app).ExpectCode(http.StatusFound)

	require.Equal(t, "/v0/characters/8", resp.Header.Get("Location"))
}

func TestHandler_GetCharacter_Redirect_cached(t *testing.T) {
	t.Parallel()
	c := cache.NewMemoryCache()
	require.NoError(t,
		c.Set(context.Background(), cachekey.Character(7), res.CharacterV0{Redirect: 8}, time.Hour))

	app := test.GetWebApp(t, test.Mock{Cache: c})

	resp := test.New(t).Get("/v0/characters/7").Execute(app).ExpectCode(http.StatusFound)

	require.Equal(t, "/v0/characters/8", resp.Header.Get("Location"))
}

func TestHandler_GetCharacter_NSFW(t *testing.T) {
	t.Parallel()
	m := mocks.NewCharacterRepo(t)
	m.EXPECT().Get(mock.Anything, model.CharacterIDType(7)).Return(model.Character{ID: 7, NSFW: true}, nil)

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(domain.Auth{ID: 1, RegTime: time.Unix(1e9, 0)}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(domain.Permission{}, nil)

	app := test.GetWebApp(t, test.Mock{
		CharacterRepo: m,
		AuthRepo:      mockAuth,
	})

	var r res.CharacterV0
	resp := test.New(t).Get("/v0/characters/7").Header(fiber.HeaderAuthorization, "Bearer v").
		Execute(app).
		JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode, resp.BodyString())
	require.Equal(t, model.CharacterIDType(7), r.ID)
}

func TestHandler_GetCharacterImage_200(t *testing.T) {
	t.Parallel()
	m := mocks.NewCharacterRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Character{ID: 7, Image: "temp"}, nil)

	app := test.GetWebApp(t, test.Mock{CharacterRepo: m})

	for _, imageType := range []string{"large", "grid", "medium", "small"} {
		t.Run(imageType, func(t *testing.T) {
			t.Parallel()

			resp := test.New(t).Get("/v0/characters/7/image?type=" + imageType).Execute(app)
			require.Equal(t, http.StatusFound, resp.StatusCode, resp.BodyString())
		})
	}
}

func TestHandler_GetCharacterImage_400(t *testing.T) {
	t.Parallel()
	m := mocks.NewCharacterRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Character{ID: 7, Image: "temp"}, nil)

	app := test.GetWebApp(t, test.Mock{CharacterRepo: m})

	resp := test.New(t).Get("/v0/characters/7/image").Execute(app)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, resp.BodyString())
}
