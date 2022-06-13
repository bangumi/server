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
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/internal/web/res"
)

func TestHappyPath(t *testing.T) {
	t.Parallel()
	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Subject{ID: 7}, nil)

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(domain.Auth{RegTime: time.Unix(1e10, 0)}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(domain.Permission{}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo:    mockAuth,
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

	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Subject{NSFW: true}, nil)

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(domain.Auth{ID: 1, RegTime: time.Unix(1e9, 0)}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(domain.Permission{}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo:    mockAuth,
			SubjectRepo: m,
		},
	)

	resp := test.New(t).Get("/v0/subjects/7").Header(fiber.HeaderAuthorization, "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode, "200 for authorized user")
}

func TestNSFW_404(t *testing.T) {
	t.Parallel()

	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Subject{NSFW: true}, nil)

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(domain.Permission{}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo:    mockAuth,
			SubjectRepo: m,
		},
	)

	resp := test.New(t).Get("/v0/subjects/7").Header("authorization", "Bearer token").
		Execute(app)

	require.Equal(t, http.StatusNotFound, resp.StatusCode, "404 for unauthorized user")
}

func Test_web_subject_Redirect(t *testing.T) {
	t.Parallel()
	m := mocks.NewSubjectRepo(t)
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
	m := mocks.NewSubjectRepo(t)

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

func TestHandler_GetSubjectImage_302(t *testing.T) {
	t.Parallel()
	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, mock.Anything).Return(model.Subject{Image: "temp"}, nil)

	app := test.GetWebApp(t, test.Mock{SubjectRepo: m})

	for _, imageType := range []string{"small", "grid", "large", "medium", "common"} {
		t.Run(imageType, func(t *testing.T) {
			t.Parallel()

			resp := test.New(t).Get("/v0/subjects/1/image?type=" + imageType).Execute(app)
			require.Equal(t, http.StatusFound, resp.StatusCode, resp.BodyString())
		})
	}
}

func TestHandler_GetSubjectImage_400(t *testing.T) {
	t.Parallel()
	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, mock.Anything).Return(model.Subject{Image: "temp"}, nil)

	app := test.GetWebApp(t, test.Mock{SubjectRepo: m})

	resp := test.New(t).Get("/v0/subjects/1/image").Execute(app)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, resp.BodyString())
}
