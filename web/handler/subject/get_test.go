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

package subject_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/tag"
	"github.com/bangumi/server/web/accessor"
	subjectHandler "github.com/bangumi/server/web/handler/subject"
	"github.com/bangumi/server/web/internal/ctxkey"
	"github.com/bangumi/server/web/res"
)

func TestSubject_Get(t *testing.T) {
	t.Parallel()
	var subjectID model.SubjectID = 7

	e := echo.New()

	g := e.Group("", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(ctxkey.User, &accessor.Accessor{Auth: auth.Auth{Login: true, RegTime: time.Time{}, ID: 1}, Login: true})
			return next(c)
		}
	})

	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, subjectID, mock.Anything).Return(model.Subject{ID: subjectID}, nil)

	ep := mocks.NewEpisodeRepo(t)
	ep.EXPECT().Count(mock.Anything, subjectID, mock.Anything).Return(3, nil)

	tagRepo := mocks.NewTagRepo(t)
	tagRepo.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return([]tag.Tag{}, nil)

	s, err := subjectHandler.New(nil, m, nil, nil, ep, tagRepo)
	require.NoError(t, err)
	s.Routes(g)

	var r res.SubjectV0
	htest.New(t, e).
		Header(echo.HeaderAuthorization, "Bearer token").
		Get("/subjects/7").
		JSON(&r).
		ExpectCode(http.StatusOK)

	require.EqualValues(t, 7, r.ID)
}

func TestSubject_Get_Redirect(t *testing.T) {
	t.Parallel()
	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, model.SubjectID(8), mock.Anything).Return(model.Subject{Redirect: 2}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			SubjectRepo: m,
		},
	)

	resp := htest.New(t, app).Get("/v0/subjects/8")

	require.Equal(t, http.StatusFound, resp.StatusCode, "302 for redirect repository")
	require.Equal(t, "/v0/subjects/2", resp.Header.Get("location"))
}

func TestSubject_Get_NSFW_200(t *testing.T) {
	t.Parallel()

	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, model.SubjectID(7), mock.Anything).Return(model.Subject{NSFW: true}, nil)

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(auth.UserInfo{ID: 1, RegTime: time.Unix(1e9, 0)}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).
		Return(auth.Permission{}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo:    mockAuth,
			SubjectRepo: m,
		},
	)

	resp := htest.New(t, app).Header(echo.HeaderAuthorization, "Bearer token").Get("/v0/subjects/7")

	require.Equal(t, http.StatusOK, resp.StatusCode, "200 for authorized user")
}

func TestSubject_Get_NSFW_404(t *testing.T) {
	t.Parallel()

	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, model.SubjectID(7), subject.Filter{NSFW: null.NewBool(false)}).
		Return(model.Subject{}, gerr.ErrSubjectNotFound)

	app := test.GetWebApp(t,
		test.Mock{SubjectRepo: m},
	)

	resp := htest.New(t, app).Get("/v0/subjects/7")

	require.Equal(t, http.StatusNotFound, resp.StatusCode, "404 for unauthorized user")
}

func TestSubject_Get_bad_id(t *testing.T) {
	t.Parallel()
	m := mocks.NewSubjectRepo(t)

	app := test.GetWebApp(t, test.Mock{SubjectRepo: m})

	for _, path := range []string{"/v0/subjects/0", "/v0/subjects/-1", "/v0/subjects/a"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			resp := htest.New(t, app).Get(path)

			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "400 for redirect subject id")
		})
	}
}

func TestSubject_GetImage_302(t *testing.T) {
	t.Parallel()
	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, model.SubjectID(1), mock.Anything).Return(model.Subject{ID: 1, Image: "temp"}, nil)
	m.EXPECT().Get(mock.Anything, model.SubjectID(3), mock.Anything).Return(model.Subject{ID: 1}, nil)

	app := test.GetWebApp(t, test.Mock{SubjectRepo: m})

	for _, imageType := range []string{"small", "grid", "large", "medium", "common"} {
		t.Run(imageType, func(t *testing.T) {
			t.Parallel()

			resp := htest.New(t, app).Get("/v0/subjects/1/image?type=" + imageType)
			require.Equal(t, http.StatusFound, resp.StatusCode, resp.BodyString())
			expected, _ := res.SubjectImage("temp").Select(imageType)
			require.Equal(t, expected, resp.Header.Get("Location"), "expect redirect to image url")

			// should redirect to default image
			resp = htest.New(t, app).Get("/v0/subjects/3/image?type=" + imageType)
			require.Equal(t, http.StatusFound, resp.StatusCode, resp.BodyString())
			require.Equal(t, res.DefaultImageURL, resp.Header.Get("Location"), "should redirect to default image")
		})
	}
}

func TestSubject_GetImage_400(t *testing.T) {
	t.Parallel()
	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(model.Subject{Image: "temp"}, nil)

	app := test.GetWebApp(t, test.Mock{SubjectRepo: m})

	htest.New(t, app).Get("/v0/subjects/1/image").ExpectCode(http.StatusBadRequest)
}
