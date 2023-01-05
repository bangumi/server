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

package ctrl

import (
	"context"

	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/null"
)

func (ctl Ctrl) CountEpisode(ctx context.Context, subjectID model.SubjectID, epType *episode.Type) (int64, error) {
	count, err := ctl.episode.Count(ctx, subjectID, episode.Filter{Type: null.NewFromPtr(epType)})
	if err != nil {
		return 0, errgo.Wrap(err, "episode.Count")
	}

	return count, nil
}
