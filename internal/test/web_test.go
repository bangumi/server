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

package test_test

import (
	"testing"

	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/test"
)

func TestGetWebApp(t *testing.T) {
	t.Parallel()

	test.GetWebApp(t,
		test.Mock{
			SubjectRepo: mocks.NewSubjectRepo(t),
			AuthRepo:    mocks.NewAuthRepo(t),
			EpisodeRepo: mocks.NewEpisodeRepo(t),
			CommentRepo: mocks.NewCommentRepo(t),
			TopicRepo:   mocks.NewTopicRepo(t),
			Cache:       mocks.NewCache(t),
		},
	)

	test.GetWebApp(t,
		test.Mock{
			SubjectRepo: mocks.NewSubjectRepo(t),
			AuthRepo:    mocks.NewAuthRepo(t),
			EpisodeRepo: mocks.NewEpisodeRepo(t),
			Cache:       mocks.NewCache(t),
		},
	)
}
