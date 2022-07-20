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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/web/res"
)

func TestHandler_GetEpisode(t *testing.T) {
	t.Parallel()
	m := mocks.NewEpisodeRepo(t)
	m.EXPECT().Get(mock.Anything, model.EpisodeID(7)).Return(model.Episode{ID: 7, SubjectID: 3}, nil)
	s := mocks.NewSubjectRepo(t)
	s.EXPECT().Get(mock.Anything, model.SubjectID(3)).Return(model.Subject{ID: 3}, nil)

	app := test.GetWebApp(t, test.Mock{EpisodeRepo: m, SubjectRepo: s})

	var episode res.Episode
	resp := test.New(t).Get("/v0/episodes/7").Execute(app).JSON(&episode)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.EqualValues(t, 7, episode.ID)
}
