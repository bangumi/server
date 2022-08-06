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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestUser_PatchEpisodeCollectionBatch(t *testing.T) {
	t.Parallel()
	const sid model.SubjectID = 8
	const uid model.UserID = 1

	var eIDs []model.EpisodeID
	var eType model.EpisodeCollection

	a := mocks.NewAuthService(t)
	a.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: uid}, nil)

	e := mocks.NewEpisodeRepo(t)
	e.EXPECT().WithQuery(mock.Anything).Return(e)
	e.EXPECT().Count(mock.Anything, mock.Anything).Return(4, nil)
	e.EXPECT().List(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]model.Episode{
		{ID: 1},
		{ID: 2},
		{ID: 3},
		{ID: 4},
	}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().WithQuery(mock.Anything).Return(c)
	c.EXPECT().UpdateEpisodeCollection(mock.Anything, uid, sid, mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ model.UserID, _ model.SubjectID,
			episodeIDs []model.EpisodeID, collection model.EpisodeCollection) {
			eIDs = episodeIDs
			eType = collection
		}).Return(model.UserSubjectEpisodesCollection{}, nil)
	c.EXPECT().UpdateSubjectCollection(mock.Anything, uid, sid, mock.Anything, mock.Anything).Return(nil)

	app := test.GetWebApp(t, test.Mock{EpisodeRepo: e, CollectionRepo: c, AuthService: a})

	test.New(t).
		Header(fiber.HeaderAuthorization, "Bearer t").
		JSON(map[string]any{
			"episode_id": []int{1, 2, 3},
			"type":       model.EpisodeCollectionDone,
		}).
		Patch(fmt.Sprintf("/v0/users/-/collections/%d/episodes", sid)).
		Execute(app).
		ExpectCode(http.StatusNoContent)

	require.Equal(t, []model.EpisodeID{1, 2, 3}, eIDs)
	require.Equal(t, model.EpisodeCollectionDone, eType)
}
