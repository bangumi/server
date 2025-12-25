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
	"github.com/trim21/htest"

	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func TestHandler_GetEpisode(t *testing.T) {
	t.Parallel()
	m := mocks.NewEpisodeRepo(t)
	m.EXPECT().Get(mock.Anything, model.EpisodeID(7)).Return(episode.Episode{ID: 7, SubjectID: 3}, nil)
	s := mocks.NewSubjectRepo(t)
	s.EXPECT().Get(mock.Anything, model.SubjectID(3), mock.Anything).Return(model.Subject{ID: 3}, nil)

	app := test.GetWebApp(t, test.Mock{EpisodeRepo: m, SubjectRepo: s})

	var e res.Episode
	htest.New(t, app).
		Get("/v0/episodes/7").
		JSON(&e).
		ExpectCode(http.StatusOK)

	require.EqualValues(t, 7, e.ID)
}

func TestHandler_ListEpisodeReverse(t *testing.T) {
	t.Parallel()

	subjectID := model.SubjectID(42)
	episodes := []episode.Episode{{ID: 10, SubjectID: subjectID}}

	m := mocks.NewEpisodeRepo(t)
	s := mocks.NewSubjectRepo(t)

	s.EXPECT().Get(mock.Anything, subjectID, mock.Anything).Return(model.Subject{ID: subjectID}, nil)

	filterMatcher := mock.MatchedBy(func(f episode.Filter) bool {
		return f.Reverse && !f.Type.Set
	})

	m.EXPECT().Count(mock.Anything, subjectID, filterMatcher).Return(int64(len(episodes)), nil)
	m.EXPECT().List(mock.Anything, subjectID, filterMatcher, req.EpisodeDefaultLimit, 0).Return(episodes, nil)

	app := test.GetWebApp(t, test.Mock{EpisodeRepo: m, SubjectRepo: s})

	var resp res.PagedG[res.Episode]
	htest.New(t, app).
		Get("/v0/episodes?subject_id=42&reverse=1").
		JSON(&resp).
		ExpectCode(http.StatusOK)

	require.EqualValues(t, 1, resp.Total)
	require.Len(t, resp.Data, 1)
	require.EqualValues(t, 10, resp.Data[0].ID)
}

func TestHandler_ListEpisodeReverseInvalid(t *testing.T) {
	t.Parallel()

	subjectID := model.SubjectID(42)

	m := mocks.NewEpisodeRepo(t)
	s := mocks.NewSubjectRepo(t)

	s.EXPECT().Get(mock.Anything, subjectID, mock.Anything).Return(model.Subject{ID: subjectID}, nil)

	app := test.GetWebApp(t, test.Mock{EpisodeRepo: m, SubjectRepo: s})

	htest.New(t, app).
		Get("/v0/episodes?subject_id=42&reverse=notabool").
		ExpectCode(http.StatusBadRequest)
}
