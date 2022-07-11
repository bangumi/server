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

	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestHandler_GetIndex_HappyPath(t *testing.T) {
	t.Parallel()
	m := mocks.NewIndexRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m})

	resp := test.New(t).Get("/v0/indices/7").Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetIndex_NSFW(t *testing.T) {
	t.Parallel()
	m := mocks.NewIndexRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Index{ID: 7, NSFW: true}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: m})

	resp := test.New(t).Get("/v0/indices/7").Execute(app)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
