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

package req_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/req"
)

func TestEpisodeCollection(t *testing.T) {
	t.Parallel()
	v := validator.New()

	require.NoError(t, v.RegisterValidation(req.EpisodeCollectionTagName, req.EpisodeCollection))

	require.Error(t, v.Var(model.EpisodeCollectionType(4), req.EpisodeCollectionTagName))
	require.Error(t, v.Var(model.EpisodeCollectionType(0), req.EpisodeCollectionTagName+",required"))
	require.NoError(t, v.Var(model.EpisodeCollectionDone, req.EpisodeCollectionTagName))
	require.NoError(t, v.Var(model.EpisodeCollectionDropped, req.EpisodeCollectionTagName))
	require.NoError(t, v.Var(model.EpisodeCollectionWish, req.EpisodeCollectionTagName))
}
