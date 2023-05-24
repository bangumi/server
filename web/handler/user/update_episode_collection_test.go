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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestUser_PatchEpisodeCollectionBatch(t *testing.T) {
	t.Parallel()
	const sid model.SubjectID = 8
	const uid model.UserID = 1
	subject := model.Subject{ID: sid, TypeID: model.SubjectTypeAll}

	var eIDs []model.EpisodeID
	var eType collection.EpisodeCollection

	a := mocks.NewAuthService(t)
	a.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.Auth{ID: uid}, nil)

	e := mocks.NewEpisodeRepo(t)
	e.EXPECT().List(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]episode.Episode{
		{ID: 1},
		{ID: 2},
		{ID: 3},
		{ID: 4},
	}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().WithQuery(mock.Anything).Return(c)
	c.EXPECT().UpdateEpisodeCollection(mock.Anything, uid, sid, mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ model.UserID, _ model.SubjectID,
			episodeIDs []model.EpisodeID, collection collection.EpisodeCollection, _ time.Time) {
			eIDs = episodeIDs
			eType = collection
		}).Return(collection.UserSubjectEpisodesCollection{}, nil)
	c.EXPECT().UpdateSubjectCollection(mock.Anything, uid, subject, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	c.EXPECT().GetSubjectCollection(mock.Anything, uid, sid).Return(collection.UserSubjectCollection{SubjectID: sid}, nil)

	app := test.GetWebApp(t, test.Mock{EpisodeRepo: e, CollectionRepo: c, AuthService: a})

	htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer t").
		BodyJSON(map[string]any{
			"episode_id": []int{1, 2, 3},
			"type":       collection.EpisodeCollectionDone,
		}).
		Patch(fmt.Sprintf("/v0/users/-/collections/%d/episodes", sid)).
		ExpectCode(http.StatusNoContent)

	require.Equal(t, []model.EpisodeID{1, 2, 3}, eIDs)
	require.Equal(t, collection.EpisodeCollectionDone, eType)
}

func TestUser_PutEpisodeCollection(t *testing.T) {
	t.Parallel()
	const sid model.SubjectID = 8
	const eid model.EpisodeID = 10
	const uid model.UserID = 1
	subject := model.Subject{ID: sid, TypeID: model.SubjectTypeAll}

	var eIDs []model.EpisodeID
	var eType collection.EpisodeCollection

	a := mocks.NewAuthService(t)
	a.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.Auth{ID: uid}, nil)

	e := mocks.NewEpisodeRepo(t)
	e.EXPECT().Get(mock.Anything, eid).Return(episode.Episode{ID: eid, SubjectID: sid}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().WithQuery(mock.Anything).Return(c)
	c.EXPECT().GetSubjectCollection(mock.Anything, uid, sid).Return(collection.UserSubjectCollection{SubjectID: sid}, nil)
	c.EXPECT().UpdateEpisodeCollection(mock.Anything, uid, sid, mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ model.UserID, _ model.SubjectID,
			episodeIDs []model.EpisodeID, collection collection.EpisodeCollection, _ time.Time) {
			eIDs = episodeIDs
			eType = collection
		}).Return(collection.UserSubjectEpisodesCollection{}, nil)
	c.EXPECT().UpdateSubjectCollection(mock.Anything, uid, subject, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	app := test.GetWebApp(t, test.Mock{EpisodeRepo: e, CollectionRepo: c, AuthService: a})

	htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer t").
		BodyJSON(map[string]any{"type": collection.EpisodeCollectionDone}).
		Put(fmt.Sprintf("/v0/users/-/collections/-/episodes/%d", eid)).
		ExpectCode(http.StatusNoContent)

	require.Equal(t, []model.EpisodeID{eid}, eIDs)
	require.Equal(t, collection.EpisodeCollectionDone, eType)
}
