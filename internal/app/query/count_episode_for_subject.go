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

package query

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/app/internal/cachekey"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func (q Query) CountEpisode(ctx context.Context, subjectID model.SubjectID, epType *model.EpType) (int64, error) {
	key := cachekey.EpisodeCount(subjectID, epType)
	var count int64
	ok, err := q.cache.Get(ctx, key, &ctx)
	if err != nil {
		return 0, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		return count, nil
	}

	if epType == nil {
		count, err = q.episode.Count(ctx, subjectID)
		if err != nil {
			return 0, errgo.Wrap(err, "episode.Count")
		}
	} else {
		count, err = q.episode.CountByType(ctx, subjectID, *epType)
		if err != nil {
			return 0, errgo.Wrap(err, "episode.CountByType")
		}
	}

	if err := q.cache.Set(ctx, key, count, time.Hour); err != nil {
		q.log.Error("failed to set cache", zap.Error(err))
	}

	return count, nil
}
