// SPDX-License-Identifier: AGPL-3.0-only
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package res_test

import (
	"testing"

	"github.com/bytedance/sonic"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/web/res"
)

func TestToSubjectV0_TagTotalCount(t *testing.T) {
	t.Parallel()

	subject := model.Subject{
		Tags: []model.Tag{{Name: "tag", Count: 10, TotalCount: 100}},
	}
	want := []res.SubjectTag{{Name: "tag", Count: 10, TotalCount: 100}}

	require.Equal(t, want, res.ToSubjectV0(subject, 0, nil).Tags)
}

func TestToSlimSubjectV0_TagTotalCount(t *testing.T) {
	t.Parallel()

	subject := model.Subject{
		Tags: []model.Tag{{Name: "tag", Count: 10, TotalCount: 100}},
	}
	want := []res.SubjectTag{{Name: "tag", Count: 10, TotalCount: 100}}

	require.Equal(t, want, res.ToSlimSubjectV0(subject).Tags)
}

func TestSubjectTag_JSON(t *testing.T) {
	t.Parallel()

	data, err := sonic.Marshal(res.SubjectTag{Name: "tag", Count: 10, TotalCount: 100})
	require.NoError(t, err)
	require.JSONEq(t, `{"name":"tag","count":10,"total_count":100}`, string(data))
	require.NotContains(t, string(data), "total_cont")
}
