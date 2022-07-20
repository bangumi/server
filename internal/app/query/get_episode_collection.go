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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func (q Query) GetUserEpisodeCollection(
	ctx context.Context, user domain.Auth, episodeID model.EpisodeID) (model.UserEpisodeCollection, model.Episode, error) {
	e, err := q.GetEpisode(ctx, episodeID)
	if err != nil {
		return model.UserEpisodeCollection{}, e, err
	}

	m, err := q.collection.GetSubjectEpisodesCollection(ctx, user.ID, e.SubjectID)
	if err != nil {
		return model.UserEpisodeCollection{}, e, errgo.Wrap(err, "collectionRepo.GetSubjectEpisodesCollection")
	}

	ec, ok := m[episodeID]
	if ok {
		return ec, e, nil
	}

	return model.UserEpisodeCollection{ID: episodeID}, e, nil
}
