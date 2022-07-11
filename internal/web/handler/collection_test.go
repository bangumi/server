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
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/web/res"
)

func TestHandler_GetCollection(t *testing.T) {
	t.Parallel()
	const username = "ni"
	const userID model.UserID = 7
	const subjectID model.SubjectID = 9

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, username).Return(model.User{ID: userID, UserName: username}, nil)
	c := mocks.NewCollectionRepo(t)
	c.EXPECT().GetSubjectCollection(mock.Anything, userID, mock.Anything).
		Return(model.SubjectCollection{SubjectID: subjectID}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: m, CollectionRepo: c})

	var r res.SubjectCollection
	resp := test.New(t).Get(fmt.Sprintf("/v0/users/%s/collections/%d", username, subjectID)).
		Execute(app).
		JSON(&r).
		ExpectCode(http.StatusOK)

	require.Equal(t, subjectID, r.SubjectID, resp.BodyString())
}

func TestHandler_GetCollection_other_user(t *testing.T) {
	t.Parallel()
	const username = "ni"
	const userID model.UserID = 7
	const subjectID model.SubjectID = 9

	a := mocks.NewAuthService(t)
	a.EXPECT().GetByTokenWithCache(mock.Anything, "v").Return(domain.Auth{ID: userID + 1}, nil)

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, username).Return(model.User{ID: userID, UserName: username}, nil)
	c := mocks.NewCollectionRepo(t)
	c.EXPECT().GetSubjectCollection(mock.Anything, userID, mock.Anything).
		Return(model.SubjectCollection{SubjectID: subjectID, Private: true}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: m, AuthService: a, CollectionRepo: c})

	resp := test.New(t).Get(fmt.Sprintf("/v0/users/%s/collections/%d", username, subjectID)).
		Header(fiber.HeaderAuthorization, "Bearer v").
		Execute(app)

	require.Equal(t, http.StatusNotFound, resp.StatusCode, resp.BodyString())
}
