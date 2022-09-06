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

package model_test

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/model"
)

func TestEpisodeCollection_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Raw      []byte
		Expected model.EpisodeCollection
		Err      bool
	}{
		{Raw: []byte("1"), Expected: 1},
		{Raw: []byte("3"), Expected: 3},
		{Raw: []byte("0"), Err: true},
	}

	for _, tc := range testCases {
		var r model.EpisodeCollection
		if tc.Err {
			require.Error(t, json.Unmarshal(tc.Raw, &r))
			continue
		}

		require.NoError(t, json.Unmarshal(tc.Raw, &r))
		require.Equal(t, tc.Expected, r)
	}
}

func TestSubjectCollection_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Raw      []byte
		Expected model.SubjectCollection
		Err      bool
	}{
		{Raw: []byte("1"), Expected: 1},
		{Raw: []byte("3"), Expected: 3},
		{Raw: []byte("0"), Err: true},
	}

	for _, tc := range testCases {
		var r model.SubjectCollection
		if tc.Err {
			require.Error(t, json.Unmarshal(tc.Raw, &r))
			continue
		}

		require.NoError(t, json.Unmarshal(tc.Raw, &r))
		require.Equal(t, tc.Expected, r)
	}
}
