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

package character_test

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
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/web/res"
)

func TestCharacter_Get(t *testing.T) {
	t.Parallel()
	m := mocks.NewCharacterRepo(t)
	m.EXPECT().Get(mock.Anything, model.CharacterID(7)).Return(model.Character{ID: 7}, nil)

	app := test.GetWebApp(t, test.Mock{CharacterRepo: m})

	var r res.CharacterV0
	test.New(t).Get("/v0/characters/7").
		Execute(app).
		JSON(&r).
		ExpectCode(http.StatusOK)
	require.EqualValues(t, 7, r.ID)
}

func TestCharacter_Get_redirect(t *testing.T) {
	t.Parallel()
	m := mocks.NewCharacterRepo(t)
	m.EXPECT().Get(mock.Anything, model.CharacterID(7)).Return(model.Character{ID: 7, Redirect: 8}, nil)

	app := test.GetWebApp(t, test.Mock{CharacterRepo: m})

	resp := test.New(t).Get("/v0/characters/7").Execute(app).ExpectCode(http.StatusFound)

	require.Equal(t, "/v0/characters/8", resp.Header.Get("Location"))
}

func TestCharacter_Get_NSFW(t *testing.T) {
	t.Parallel()
	m := mocks.NewCharacterRepo(t)
	m.EXPECT().Get(mock.Anything, model.CharacterID(7)).Return(model.Character{ID: 7, NSFW: true}, nil)

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).
		Return(domain.AuthUserInfo{ID: 1, RegTime: time.Unix(1e9, 0)}, nil)
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
	require.Equal(t, model.CharacterID(7), r.ID)
}

func TestCharacter_GetImage(t *testing.T) {
	t.Parallel()
	m := mocks.NewCharacterRepo(t)
	m.EXPECT().Get(mock.Anything, model.CharacterID(7)).Return(model.Character{ID: 7, Image: "temp"}, nil)
	m.EXPECT().Get(mock.Anything, model.CharacterID(8)).Return(model.Character{ID: 8}, nil)

	app := test.GetWebApp(t, test.Mock{CharacterRepo: m})

	for _, imageType := range []string{"large", "grid", "medium", "small"} {
		imageType := imageType
		t.Run(imageType, func(t *testing.T) {
			t.Parallel()

			resp := test.New(t).Get("/v0/characters/7/image?type=" + imageType).Execute(app)
			require.Equal(t, http.StatusFound, resp.StatusCode, resp.BodyString())
			expected, _ := res.PersonImage("temp").Select(imageType)
			require.Equal(t, expected, resp.Header.Get("Location"), "expect redirect to image url")

			// should redirect to default image
			resp = test.New(t).Get("/v0/characters/8/image?type=" + imageType).Execute(app)
			require.Equal(t, http.StatusFound, resp.StatusCode, resp.BodyString())
			require.Equal(t, res.DefaultImageURL, resp.Header.Get("Location"), "should redirect to default image")
		})
	}
}

func TestCharacter_GetImage_400(t *testing.T) {
	t.Parallel()
	m := mocks.NewCharacterRepo(t)
	m.EXPECT().Get(mock.Anything, model.CharacterID(7)).Return(model.Character{ID: 7, Image: "temp"}, nil)

	app := test.GetWebApp(t, test.Mock{CharacterRepo: m})

	resp := test.New(t).Get("/v0/characters/7/image").Execute(app)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, resp.BodyString())
}
