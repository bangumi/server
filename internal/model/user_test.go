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

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/model"
)

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	var s = model.UserPrivacySettings{}

	s.Unmarshal([]byte("a:4:{i:1;i:2;i:30;i:2;i:20;i:2;i:21;i:0;}"))

	require.Equal(t, s.ReceiveTimelineReply, model.UserReceiveFilterNone)
}
