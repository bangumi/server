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

package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/util"
)

type Image struct {
	Cat  *int64  `php:"cat,omitempty"`
	ID   *string `php:"grp_id,omitempty"`
	Name *string `php:"grp_name,omitempty"`
}

type TimeLineImage struct {
	Cat  *int64
	ID   *string
	Name *int
}

func TestCopySameNameField(t *testing.T) {
	t.Parallel()

	var i = Image{
		Cat:  null.NewInt64(1).Ptr(),
		ID:   null.New("ii").Ptr(),
		Name: null.New("ii").Ptr(),
	}

	var dst = TimeLineImage{}

	util.CopySameNameField(&dst, &i)

	require.NotNil(t, dst.Cat)
	require.Equal(t, int64(1), *dst.Cat)

	require.NotNil(t, dst.ID)
	require.Equal(t, "ii", *dst.ID)

	require.Nil(t, dst.Name)
}
