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
)

func TestHandler_GetGroupTopic(t *testing.T) {
	t.Parallel()

	g := mocks.NewGroupRepo(t)
	g.EXPECT().GetByID(mock.Anything, model.GroupID(6)).Return(model.Group{Name: "group name", ID: 6}, nil)

	topic := mocks.NewTopicRepo(t)
	topic.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(model.Topic{ID: 1, ParentID: 6}, nil)
	topic.EXPECT().GetTopicContent(mock.Anything, mock.Anything, mock.Anything).Return(model.Comment{}, nil)
	topic.EXPECT().ListReplies(mock.Anything, mock.Anything, model.TopicID(1), 0, 0).Return([]model.Comment{}, nil)

	app := test.GetWebApp(t, test.Mock{
		TopicRepo: topic,
		GroupRepo: g,
	})

	resp := test.New(t).Get("/p/groups/-/topics/1").Execute(app)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
