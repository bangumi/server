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

package user_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/dam"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestUser_PatchSubjectCollection(t *testing.T) {
	t.Parallel()
	const sid model.SubjectID = 8
	const uid model.UserID = 1
	subject := model.Subject{ID: sid, TypeID: model.SubjectTypeAll}

	var s = &collection.Subject{}

	a := mocks.NewAuthService(t)
	a.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.Auth{ID: uid}, nil)

	tl := mocks.NewTimeLineService(t)
	tl.EXPECT().
		ChangeSubjectCollection(mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().UpdateSubjectCollection(mock.Anything, uid, subject, mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, userID uint32,
			subject model.Subject, at time.Time, ip string,
			update func(context.Context, *collection.Subject) (*collection.Subject, error)) {
			require.Equal(t, "0.0.0.0", ip)

			s = lo.Must(update(context.Background(), s))
		}).Return(nil)

	d, err := dam.New(config.AppConfig{NsfwWord: "", DisableWords: "test_content", BannedDomain: ""})
	require.NoError(t, err)

	app := test.GetWebApp(t, test.Mock{CollectionRepo: c, AuthService: a, Dam: &d, TimeLineSrv: tl})

	htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer t").
		BodyJSON(map[string]any{
			"comment": "1 test_content 2",
			"type":    2,
			"private": true,
			"rate":    8,
			"tags":    []string{"q", "vv"},
		}).
		Patch(fmt.Sprintf("/v0/users/-/collections/%d", sid)).
		ExpectCode(http.StatusNoContent)

	require.Equal(t, collection.CollectPrivacySelf, s.Privacy())
	require.Equal(t, "1 test_content 2", s.Comment())
	require.EqualValues(t, []string{"q", "vv"}, s.Tags())
	require.EqualValues(t, 8, s.Rate())
}

func TestUser_PatchToNonExistsSubjectCollection(t *testing.T) {
	t.Parallel()
	const sid model.SubjectID = 8
	const uid model.UserID = 1
	subject := model.Subject{ID: sid, TypeID: model.SubjectTypeAll}

	a := mocks.NewAuthService(t)
	a.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.Auth{ID: uid}, nil)

	tl := mocks.NewTimeLineService(t)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().UpdateSubjectCollection(mock.Anything, uid, subject, mock.Anything, mock.Anything, mock.Anything).
		Return(gerr.ErrSubjectNotCollected)

	d, err := dam.New(config.AppConfig{NsfwWord: "", DisableWords: "test_content", BannedDomain: ""})
	require.NoError(t, err)

	app := test.GetWebApp(t, test.Mock{CollectionRepo: c, AuthService: a, Dam: &d, TimeLineSrv: tl})

	htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer t").
		BodyJSON(map[string]any{
			"comment": "1 test_content 2",
			"type":    1,
			"private": true,
			"rate":    8,
			"tags":    []string{"q", "vv"},
		}).
		Patch(fmt.Sprintf("/v0/users/-/collections/%d", sid)).
		ExpectCode(http.StatusNotFound)
}

func TestUser_PatchSubjectCollection_badID(t *testing.T) {
	t.Parallel()

	a := mocks.NewAuthService(t)
	a.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.Auth{ID: 1}, nil)

	tl := mocks.NewTimeLineService(t)
	c := mocks.NewCollectionRepo(t)

	d, err := dam.New(config.AppConfig{NsfwWord: "", DisableWords: "test_content", BannedDomain: ""})
	require.NoError(t, err)

	app := test.GetWebApp(t, test.Mock{CollectionRepo: c, AuthService: a, Dam: &d, TimeLineSrv: tl})

	badURLs := []string{
		"/v0/users/-/collections/abc",
		"/v0/users/-/collections/123_",
		"/v0/users/-/collections/s123",
		"/v0/users/-/collections/_123",
		"/v0/users/-/collections/_abc",
		"/v0/users/-/collections/_",
	}

	for _, url := range badURLs {
		htest.New(t, app).
			Header(echo.HeaderAuthorization, "Bearer t").
			BodyJSON(map[string]any{
				"comment": "1 test_content 2",
			}).
			Patch(url).
			ExpectCode(http.StatusBadRequest)
	}
}

func TestUser_PatchSubjectCollection_bad(t *testing.T) {
	t.Parallel()
	const uid model.UserID = 1
	const sid model.SubjectID = 8

	a := &mocks.AuthService{}
	a.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.Auth{ID: uid}, nil)

	t.Run("bad rate", func(t *testing.T) {
		t.Parallel()

		app := test.GetWebApp(t, test.Mock{AuthService: a})

		htest.New(t, app).
			Header(echo.HeaderAuthorization, "Bearer t").
			BodyJSON(echo.Map{"rate": 11}).
			Patch(fmt.Sprintf("/v0/users/-/collections/%d", sid)).
			ExpectCode(http.StatusBadRequest)
	})

	t.Run("bad type", func(t *testing.T) {
		t.Parallel()

		app := test.GetWebApp(t, test.Mock{AuthService: a})

		htest.New(t, app).
			Header(echo.HeaderAuthorization, "Bearer t").
			BodyJSON(echo.Map{"type": 0}).
			Patch(fmt.Sprintf("/v0/users/-/collections/%d", sid)).
			ExpectCode(http.StatusBadRequest)
	})

	t.Run("bad type", func(t *testing.T) {
		t.Parallel()

		app := test.GetWebApp(t, test.Mock{AuthService: a})

		htest.New(t, app).
			Header(echo.HeaderAuthorization, "Bearer t").
			BodyJSON(echo.Map{"tags": "vv qq"}).
			Patch(fmt.Sprintf("/v0/users/-/collections/%d", sid)).
			ExpectCode(http.StatusBadRequest)
	})

	t.Run("too long comment", func(t *testing.T) {
		t.Parallel()

		app := test.GetWebApp(t, test.Mock{AuthService: a})

		htest.New(t, app).
			Header(echo.HeaderAuthorization, "Bearer t").
			BodyJSON(echo.Map{"comment": strings.Repeat("vv qq", 200)}).
			Patch(fmt.Sprintf("/v0/users/-/collections/%d", sid)).
			ExpectCode(http.StatusBadRequest)
	})
}
