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
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/app/internal/cachekey"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func (q Query) GetEpisode(ctx context.Context, id model.EpisodeID) (model.Episode, error) {
	q.metricsEpisodeQueryCount.Inc(1)
	var key = cachekey.Episode(id)
	// try to read from cache
	var e model.Episode
	cached, err := q.cache.Get(ctx, key, &e)
	if err != nil {
		return model.Episode{}, errgo.Wrap(err, "cache.Get")
	}

	if cached {
		q.metricsEpisodeQueryCached.Inc(1)
		return e, nil
	}

	e, err = q.episode.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return model.Episode{}, domain.ErrEpisodeNotFound
		}

		return e, errgo.Wrap(err, "EpisodeRepo.Get")
	}

	if e := q.cache.Set(ctx, key, e, time.Minute); e != nil {
		q.log.Error("can't set response to cache", zap.Error(e))
	}

	return e, nil
}
