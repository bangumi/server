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

func TestUser_GetEpisodeCollection(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: 3}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().GetSubjectEpisodesCollection(mock.Anything, mock.Anything, mock.Anything).
		Return(map[model.EpisodeID]model.UserEpisodeCollection{}, nil)

	episode := mocks.NewEpisodeRepo(t)
	episode.EXPECT().Get(mock.Anything, model.EpisodeID(1)).Return(model.Episode{}, nil)

	app := test.GetWebApp(t, test.Mock{AuthService: mockAuth, CollectionRepo: c, EpisodeRepo: episode})

	var r struct {
		Type uint8 `json:"type"`
	}
	test.New(t).Get("/v0/users/-/collections/-/episodes/1").
		Header(fiber.HeaderAuthorization, "Bearer").
		Execute(app).
		JSON(&r).
		ExpectCode(http.StatusOK)
}

func TestUser_GetSubjectEpisodeCollection(t *testing.T) {
	t.Parallel()

	mockAuth := mocks.NewAuthService(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{ID: 3}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().GetSubjectEpisodesCollection(mock.Anything, mock.Anything, mock.Anything).
		Return(map[model.EpisodeID]model.UserEpisodeCollection{}, nil)

	episode := mocks.NewEpisodeRepo(t)
	episode.EXPECT().Count(mock.Anything, mock.Anything, mock.Anything).Return(20, nil)
	episode.EXPECT().List(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]model.Episode{}, nil)

	app := test.GetWebApp(t, test.Mock{AuthService: mockAuth, CollectionRepo: c, EpisodeRepo: episode})

	var r res.PagedG[struct {
		Type uint8 `json:"type"`
	}]
	test.New(t).Get("/v0/users/-/collections/8/episodes").
		Header(fiber.HeaderAuthorization, "Bearer").
		Execute(app).
		JSON(&r).
		ExpectCode(http.StatusOK)

	require.EqualValues(t, 20, r.Total)
}
