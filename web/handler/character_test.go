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
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/mocks"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/handler/cachekey"
	"github.com/bangumi/server/web/res"
)

func TestHandler_GetCharacter_HappyPath(t *testing.T) {
	t.Parallel()
	m := &mocks.CharacterRepo{}
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
	m := &mocks.CharacterRepo{}
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
	m := &mocks.CharacterRepo{}
	m.EXPECT().Get(mock.Anything, model.CharacterIDType(7)).Return(model.Character{ID: 7, NSFW: true}, nil)

	app := test.GetWebApp(t, test.Mock{
		CharacterRepo: m,
		AuthRepo:      mockAuth{domain.Auth{ID: 1, RegTime: time.Unix(1e9, 0)}},
	})

	var r res.CharacterV0
	test.New(t).Get("/v0/characters/7").Header(fiber.HeaderAuthorization, "Bearer v").
		Execute(app).
		JSON(&r).
		ExpectCode(http.StatusOK)

	require.Equal(t, model.CharacterIDType(7), r.ID)
}
