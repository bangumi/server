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

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/mocks"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/res"
)

type mockAuth struct{ u domain.Auth }

func (m mockAuth) GetByToken(ctx context.Context, token string) (domain.Auth, error) {
	return m.u, nil
}

func TestHappyPath(t *testing.T) {
	t.Parallel()
	m := &mocks.SubjectRepo{}
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Subject{ID: 7}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo:    mockAuth{domain.Auth{RegTime: time.Unix(1e10, 0)}},
			SubjectRepo: m,
		},
	)

	var r res.SubjectV0
	resp := test.New(t).Get("/v0/subjects/7").Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app).JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, uint32(7), r.ID)
}

func TestNSFW_200(t *testing.T) {
	t.Parallel()

	m := &mocks.SubjectRepo{}
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Subject{NSFW: true}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo:    mockAuth{domain.Auth{ID: 1, RegTime: time.Unix(1e9, 0)}},
			SubjectRepo: m,
		},
	)

	resp := test.New(t).Get("/v0/subjects/7").Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode, "200 for authorized user")
}

func TestNSFW_404(t *testing.T) {
	t.Parallel()

	m := &mocks.SubjectRepo{}
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Subject{NSFW: true}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo:    mockAuth{domain.Auth{}},
			SubjectRepo: m,
		},
	)

	resp := test.New(t).Get("/v0/subjects/7").Header("authorization", "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusNotFound, resp.StatusCode, "404 for unauthorized user")
}

func Test_web_subject_Redirect(t *testing.T) {
	t.Parallel()
	m := &mocks.SubjectRepo{}
	m.EXPECT().Get(mock.Anything, uint32(8)).Return(model.Subject{Redirect: 2}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			SubjectRepo: m,
		},
	)

	resp := test.New(t).Get("/v0/subjects/8").Execute(app)

	require.Equal(t, http.StatusFound, resp.StatusCode, "302 for redirect repository")
	require.Equal(t, "/v0/subjects/2", resp.Header.Get("location"))
}

func Test_web_subject_bad_id(t *testing.T) {
	t.Parallel()
	m := &mocks.SubjectRepo{}

	app := test.GetWebApp(t, test.Mock{SubjectRepo: m})

	for _, path := range []string{"/v0/subjects/0", "/v0/subjects/-1", "/v0/subjects/a"} {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			resp := test.New(t).Get(path).Execute(app)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "400 for redirect subject id")
		})
	}
}
