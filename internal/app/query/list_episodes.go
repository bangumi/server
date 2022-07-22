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

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func (q Query) ListEpisode(
	ctx context.Context,
	subjectID model.SubjectID,
	epType *model.EpType,
	limit, offset int,
) ([]model.Episode, int64, error) {
	count, err := q.CountEpisode(ctx, subjectID, epType)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return []model.Episode{}, 0, nil
	}

	if int64(offset) > count {
		return []model.Episode{}, count, ErrOffsetTooBig
	}

	if epType == nil {
		var episodes []model.Episode
		episodes, err = q.episode.List(ctx, subjectID, limit, offset)
		if err != nil {
			return nil, 0, errgo.Wrap(err, "episode.List")
		}
		return episodes, count, nil
	}

	episodes, err := q.episode.ListByType(ctx, subjectID, *epType, limit, offset)
	if err != nil {
		return nil, 0, errgo.Wrap(err, "episode.ListByType")
	}

	return episodes, count, nil
}

func (q Query) GetAllEpisodes(ctx context.Context, subjectID model.SubjectID) ([]model.Episode, error) {
	count, err := q.CountEpisode(ctx, subjectID, nil)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return []model.Episode{}, nil
	}

	var episodes []model.Episode
	episodes, err = q.episode.List(ctx, subjectID, int(count), 0)
	if err != nil {
		return nil, errgo.Wrap(err, "episode.List")
	}
	return episodes, nil
}
