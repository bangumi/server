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
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dam"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestUser_PatchSubjectCollection(t *testing.T) {
	t.Parallel()
	const sid model.SubjectID = 8
	const uid model.UserID = 1

	var call domain.SubjectCollectionUpdate

	a := mocks.NewAuthService(t)
	a.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: uid}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().GetSubjectCollection(mock.Anything, uid, mock.Anything).
		Return(model.UserSubjectCollection{}, nil)
	c.EXPECT().WithQuery(mock.Anything).Return(c)
	c.EXPECT().UpdateSubjectCollection(mock.Anything, uid, sid, mock.Anything).
		Run(func(ctx context.Context, userID model.UserID, subjectID model.SubjectID, data domain.SubjectCollectionUpdate) {
			call = data
		}).Return(nil)

	d, err := dam.New(config.AppConfig{NsfwWord: "", DisableWords: "test_content", BannedDomain: ""})
	require.NoError(t, err)

	app := test.GetWebApp(t, test.Mock{CollectionRepo: c, AuthService: a, Dam: &d})

	test.New(t).
		Header(fiber.HeaderAuthorization, "Bearer t").
		JSON(map[string]any{
			"comment": "1 test_content 2",
			"type":    1,
			"tags":    []string{"q", "vv"},
		}).
		Patch(fmt.Sprintf("/v0/users/-/collections/%d", sid)).
		Execute(app).
		ExpectCode(http.StatusNoContent)

	require.Equal(t, domain.SubjectCollectionUpdate{
		Comment: null.NewString("1 test_content 2"),
		Tags:    []string{"q", "vv"},
		Type:    null.New(model.SubjectCollection(1)),
		Privacy: null.New(model.CollectPrivacyBan),
	}, call)
}
