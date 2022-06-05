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

package domain_test

import (
	"testing"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/stretchr/testify/require"
)

func TestConvertModelCommentsToTree(t *testing.T) {
	t.Parallel()
	type args struct {
		comments []model.Comment
		related  uint32
	}

	input := []model.Comment{
		{
			Related: 0,
			ID:      1,
		},
		{
			Related: 0,
			ID:      2,
		},
		{
			Related: 1,
			ID:      3,
		},
		{
			Related: 3,
			ID:      4,
		},
	}
	want := []model.Comment{
		{
			Related: 0,
			ID:      1,
			Replies: []model.Comment{
				{
					Related: 1,
					ID:      3,
					Replies: []model.Comment{
						{
							Related: 3,
							ID:      4,
						},
					},
				},
			},
		},
		{
			Related: 0,
			ID:      2,
		},
	}

	require.Equal(t, want, domain.ConvertModelCommentsToTree(input, 0))
}
