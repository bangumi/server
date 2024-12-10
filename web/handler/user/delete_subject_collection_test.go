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
	"fmt"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/trim21/htest"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestUser_DeleteSubjectCollection(t *testing.T) {
	t.Parallel()
	const uid model.UserID = 1
	const sid model.SubjectID = 8

	a := mocks.NewAuthService(t)
	a.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.Auth{ID: uid}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().DeleteSubjectCollection(mock.Anything, uid, sid).
		Return(nil)

	app := test.GetWebApp(t, test.Mock{CollectionRepo: c, AuthService: a})

	htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer t").
		Delete(fmt.Sprintf("/v0/users/-/collections/%d", sid)).
		ExpectCode(http.StatusNoContent)
}
