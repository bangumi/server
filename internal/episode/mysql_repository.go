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
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/pkg/vars/enum"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.EpisodeRepo, error) {
	return mysqlRepo{q: q, log: log.Named("episode.mysqlRepo")}, nil
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
	if episode.Type == enum.EpTypeNormal {
		if first, err = r.firstEpisode(ctx, episode.SubjectID); err != nil {
			return model.Episode{}, err
		}
	}

	return convertDaoEpisode(episode, first), nil
}

func (r mysqlRepo) Count(ctx context.Context, subjectID model.SubjectID) (int64, error) {
	c, err := r.q.Episode.WithContext(ctx).
		Where(r.q.Episode.SubjectID.Eq(subjectID), r.q.Episode.Ban.Eq(0)).Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}

	return c, nil
}

func (r mysqlRepo) CountByType(
	ctx context.Context,
	subjectID model.SubjectID,
	epType model.EpType,
) (int64, error) {
	c, err := r.q.Episode.WithContext(ctx).
		Where(
			r.q.Episode.SubjectID.Eq(subjectID),
			r.q.Episode.Type.Eq(epType),
			r.q.Episode.Ban.Eq(0),
		).Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}

	return c, nil
}

func (r mysqlRepo) List(
	ctx context.Context, subjectID model.SubjectID, limit int, offset int,
) ([]model.Episode, error) {
	first, err := r.firstEpisode(ctx, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return []model.Episode{}, nil
		}

		return nil, err
	}

	episodes, err := r.q.Episode.WithContext(ctx).
		Where(r.q.Episode.SubjectID.Eq(subjectID), r.q.Episode.Ban.Eq(0)).
		Limit(limit).Offset(offset).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	var result = make([]model.Episode, len(episodes))
	for i, episode := range episodes {
		result[i] = convertDaoEpisode(episode, first)
	}

	return result, nil
}

func (r mysqlRepo) ListByType(
	ctx context.Context, subjectID model.SubjectID, epType model.EpType, limit int, offset int,
) ([]model.Episode, error) {
	first, err := r.firstEpisode(ctx, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return []model.Episode{}, nil
		}

		return nil, err
	}

	episodes, err := r.q.Episode.WithContext(ctx).
		Where(
			r.q.Episode.SubjectID.Eq(subjectID),
			r.q.Episode.Type.Eq(epType),
			r.q.Episode.Ban.Eq(0),
		).
		Limit(limit).Offset(offset).Find()
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
			r.q.Episode.Type.Eq(enum.EpTypeNormal),
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
	if e.Type == enum.EpTypeNormal {
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

func (r mysqlRepo) CountsBySubjectID(
	ctx context.Context, subjectID ...model.SubjectID,
) (map[model.SubjectID]int64, error) {
	var count []struct {
		SubjectID model.SubjectID
		Total     int64
	}

	err := r.q.Episode.WithContext(ctx).
		Select(r.q.Episode.SubjectID.As("SubjectID"), r.q.Episode.ID.Count().As("Total")).Group(r.q.Episode.SubjectID).
		Where(r.q.Episode.SubjectID.In(slice.ToValuer(subjectID)...)).Scan(&count)
	if err != nil {
		return nil, errgo.Wrap(err, "dal.select.count.group")
	}

	var m = make(map[model.SubjectID]int64, len(count))
	for _, s := range count {
		m[s.SubjectID] = s.Total
	}

	return m, nil
}
