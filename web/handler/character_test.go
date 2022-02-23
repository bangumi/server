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
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/handler/cachekey"
	"github.com/bangumi/server/web/res"
)

func TestHandler_GetCharacter_HappyPath(t *testing.T) {
	t.Parallel()
	m := &domain.MockCharacterRepo{}
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Character{ID: 7}, nil)

	app := test.GetWebApp(t, test.Mock{CharacterRepo: m})

	req := httptest.NewRequest(http.MethodGet, "/v0/characters/7", http.NoBody)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var r res.CharacterV0
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&r))
	require.Equal(t, uint32(7), r.ID)
}

func TestHandler_GetCharacter_Redirect(t *testing.T) {
	t.Parallel()
	m := &domain.MockCharacterRepo{}
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Character{ID: 7, Redirect: 8}, nil)

	app := test.GetWebApp(t, test.Mock{CharacterRepo: m})

	req := httptest.NewRequest(http.MethodGet, "/v0/characters/7", http.NoBody)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusFound, resp.StatusCode)
	require.Equal(t, "/v0/characters/8", resp.Header.Get("Location"))
}

func TestHandler_GetCharacter_Redirect_cached(t *testing.T) {
	t.Parallel()
	c := cache.NewMemoryCache()
	require.NoError(t,
		c.Set(context.Background(), cachekey.Character(7), res.CharacterV0{Redirect: 8}, time.Hour))

	app := test.GetWebApp(t, test.Mock{Cache: c})

	req := httptest.NewRequest(http.MethodGet, "/v0/characters/7", http.NoBody)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusFound, resp.StatusCode)
	require.Equal(t, "/v0/characters/8", resp.Header.Get("Location"))
}

func TestHandler_GetCharacter_NSFW(t *testing.T) {
	t.Parallel()
	m := &domain.MockCharacterRepo{}
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Character{ID: 7, NSFW: true}, nil)

	app := test.GetWebApp(t, test.Mock{
		CharacterRepo: m,
		AuthRepo:      mockAuth{domain.Auth{ID: 1, RegTime: time.Unix(1e9, 0)}},
	})

	req := httptest.NewRequest(http.MethodGet, "/v0/characters/7", http.NoBody)
	req.Header.Set(fiber.HeaderAuthorization, "bearer v")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var r res.CharacterV0
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&r))

	require.Equal(t, uint32(7), r.ID)
}
