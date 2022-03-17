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

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/mocks"
	"github.com/bangumi/server/model"
)

func TestHandler_GetIndex_HappyPath(t *testing.T) {
	t.Parallel()
	m := &mocks.IndexRepo{}
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m})

	req := httptest.NewRequest(http.MethodGet, "/v0/indices/7", http.NoBody)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetIndex_NSFW(t *testing.T) {
	t.Parallel()
	m := &mocks.IndexRepo{}
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7, NSFW: true}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m})

	req := httptest.NewRequest(http.MethodGet, "/v0/indices/7", http.NoBody)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
