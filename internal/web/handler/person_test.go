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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/internal/web/res"
)

func TestHandler_GetPerson_HappyPath(t *testing.T) {
	t.Parallel()
	m := mocks.NewPersonRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Person{ID: 7}, nil)

	app := test.GetWebApp(t, test.Mock{PersonRepo: m})

	resp := test.New(t).Get("/v0/persons/7").Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetPerson_Redirect(t *testing.T) {
	t.Parallel()
	m := mocks.NewPersonRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Person{ID: 7}, nil)

	app := test.GetWebApp(t, test.Mock{PersonRepo: m})

	resp := test.New(t).Get("/v0/persons/7").Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetPerson_Redirect_cached(t *testing.T) {
	t.Parallel()
	c := cache.NewMemoryCache()
	require.NoError(t,
		c.Set(context.Background(), cachekey.Person(7), res.PersonV0{Redirect: 8}, time.Hour))

	app := test.GetWebApp(t, test.Mock{Cache: c})
	resp := test.New(t).Get("/v0/persons/7").Execute(app)

	require.Equal(t, http.StatusFound, resp.StatusCode)
	require.Equal(t, "/v0/persons/8", resp.Header.Get("Location"))
}
