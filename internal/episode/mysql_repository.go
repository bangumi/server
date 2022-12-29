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

package episode

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (Repo, error) {
	return mysqlRepo{q: q, log: log.Named("episode.mysqlRepo")}, nil
}

func (r mysqlRepo) WithQuery(query *query.Query) Repo {
	return mysqlRepo{q: query, log: r.log}
}

func (r mysqlRepo) Get(ctx context.Context, episodeID model.EpisodeID) (model.Episode, error) {
	episode, err := r.q.Episode.WithContext(ctx).
		Where(r.q.Episode.ID.Eq(episodeID), r.q.Episode.Ban.Eq(0)).Limit(1).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Episode{}, domain.ErrNotFound
		}

		return model.Episode{}, errgo.Wrap(err, "dal")
	}

	var first float32
	if episode.Type == model.EpTypeNormal {
		if first, err = r.firstEpisode(ctx, episode.SubjectID); err != nil {
			return model.Episode{}, err
		}
	}

	return convertDaoEpisode(episode, first), nil
}

func (r mysqlRepo) Count(
	ctx context.Context,
	subjectID model.SubjectID,
	filter Filter,
) (int64, error) {
	q := r.q.Episode.WithContext(ctx).Where(r.q.Episode.SubjectID.Eq(subjectID), r.q.Episode.Ban.Eq(0))

	if filter.Type.Set {
		q = q.Where(r.q.Episode.Type.Eq(filter.Type.Value))
	}

	c, err := q.Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}

	return c, nil
}

func (r mysqlRepo) List(
	ctx context.Context, subjectID model.SubjectID, filter Filter, limit int, offset int,
) ([]model.Episode, error) {
	first, err := r.firstEpisode(ctx, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return []model.Episode{}, nil
		}

		return nil, err
	}

	q := r.q.Episode.WithContext(ctx).Where(r.q.Episode.SubjectID.Eq(subjectID), r.q.Episode.Ban.Eq(0))
	if filter.Type.Set {
		q = q.Where(r.q.Episode.Type.Eq(filter.Type.Value))
	}

	episodes, err := q.Limit(limit).Offset(offset).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	var result = make([]model.Episode, len(episodes))
	for i, episode := range episodes {
		result[i] = convertDaoEpisode(episode, first)
	}

	return result, nil
}

func (r mysqlRepo) firstEpisode(ctx context.Context, subjectID model.SubjectID) (float32, error) {
	episode, err := r.q.Episode.WithContext(ctx).
		Where(
			r.q.Episode.SubjectID.Eq(subjectID),
			r.q.Episode.Type.Eq(model.EpTypeNormal),
			r.q.Episode.Ban.Eq(0),
		).
		Order(r.q.Episode.Disc, r.q.Episode.Sort).Limit(1).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}

		return 0, errgo.Wrap(err, "dal")
	}

	return episode.Sort, nil
}

func convertDaoEpisode(e *dao.Episode, firstEpisode float32) model.Episode {
	var ep float32
	if e.Type == model.EpTypeNormal {
		ep = e.Sort - firstEpisode + 1
	}

	return model.Episode{
		Airdate:     e.Airdate,
		Name:        e.Name,
		NameCN:      e.NameCn,
		SubjectID:   e.SubjectID,
		Duration:    e.Duration,
		Description: e.Desc,
		Type:        e.Type,
		Sort:        e.Sort,
		Ep:          ep,
		Comment:     e.Comment,
		Disc:        e.Disc,
		ID:          e.ID,
	}
}
