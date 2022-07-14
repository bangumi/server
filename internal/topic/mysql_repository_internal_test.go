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

package topic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/model"
)

func TestMysqlRepo_convertDao(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in interface{}
		id model.TopicID
	}{
		{
			in: &dao.GroupTopic{ID: 1},
			id: 1,
		},
		{
			in: &dao.SubjectTopic{ID: 2},
			id: 2,
		},
	}
	for _, tt := range tests {
		p, err := convertDao(tt.in)
		require.NoError(t, err)
		require.Equal(t, p.ID, tt.id)
	}
}

//nolint:funlen
func TestMysqlRepo_convertModelTopics(t *testing.T) {
	t.Parallel()

	ti := time.Now()
	tif := time.Unix(ti.Unix(), 0)

	tests := []struct {
		in   interface{}
		want []model.Topic
	}{
		{
			in: []*dao.SubjectTopic{
				{
					ID:          3,
					SubjectID:   3,
					UID:         123,
					Title:       "321",
					CreatedTime: uint32(ti.Unix()),
					UpdatedTime: uint32(ti.Unix()),
					Replies:     12,
					State:       0,
					Status:      0,
				},
			},
			want: []model.Topic{
				{
					CreatedTime: tif,
					UpdatedTime: tif,
					Title:       "321",
					ID:          3,
					UID:         123,
					Replies:     12,
					ObjectID:    3,
					State:       0,
					Status:      0,
				},
			},
		},
		{
			in: []*dao.GroupTopic{
				{
					ID:          3,
					GroupID:     3,
					UID:         123,
					Title:       "321",
					CreatedTime: uint32(ti.Unix()),
					UpdatedTime: uint32(ti.Unix()),
					Replies:     12,
					State:       0,
					Status:      0,
				},
			},
			want: []model.Topic{
				{
					CreatedTime: tif,
					UpdatedTime: tif,
					Title:       "321",
					ID:          3,
					UID:         123,
					Replies:     12,
					ObjectID:    3,
					State:       0,
					Status:      0,
				},
			},
		},
	}

	for _, tt := range tests {
		topics := convertModelTopics(tt.in)
		require.Equal(t, topics, tt.want, "convert topics")
	}
}
