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

//go:build !gen

package episode

import (
	"context"
	"errors"
	"time"

	"github.com/trim21/errgo"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
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

func (r mysqlRepo) Get(ctx context.Context, episodeID model.EpisodeID) (Episode, error) {
	episode, err := r.q.Episode.WithContext(ctx).
		Where(r.q.Episode.ID.Eq(episodeID), r.q.Episode.Ban.Eq(0)).Limit(1).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Episode{}, gerr.ErrNotFound
		}

		return Episode{}, errgo.Wrap(err, "dal")
	}

	var first float32
	if episode.Type == TypeNormal {
		if first, err = r.firstEpisode(ctx, episode.SubjectID); err != nil {
			return Episode{}, err
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
) ([]Episode, error) {
	first, err := r.firstEpisode(ctx, subjectID)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return []Episode{}, nil
		}

		return nil, err
	}

	q := r.q.Episode.WithContext(ctx).Where(r.q.Episode.SubjectID.Eq(subjectID), r.q.Episode.Ban.Eq(0))
	if filter.Type.Set {
		q = q.Where(r.q.Episode.Type.Eq(filter.Type.Value))
	}

	episodes, err := q.Order(r.q.Episode.Disc, r.q.Episode.Type, r.q.Episode.Sort).Limit(limit).Offset(offset).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	var result = make([]Episode, len(episodes))
	for i, episode := range episodes {
		result[i] = convertDaoEpisode(episode, first)
	}

	return result, nil
}

func (r mysqlRepo) firstEpisode(ctx context.Context, subjectID model.SubjectID) (float32, error) {
	episode, err := r.q.Episode.WithContext(ctx).
		Where(
			r.q.Episode.SubjectID.Eq(subjectID),
			r.q.Episode.Type.Eq(TypeNormal),
			r.q.Episode.Ban.Eq(0),
		).
		Order(r.q.Episode.Disc, r.q.Episode.Sort).Limit(1).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}

		return 0, errgo.Wrap(err, "dal")
	}

	return episode.Sort, nil
}

func convertDaoEpisode(e *dao.Episode, firstEpisode float32) Episode {
	var ep float32
	if e.Type == TypeNormal {
		ep = e.Sort - firstEpisode + 1
	}

	return Episode{
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

func (r mysqlRepo) GetAllComment(ctx context.Context, episodeID model.EpisodeID, offset int, limit int) ([]model.EpisodeComment, error) {
	s, err := r.q.EpisodeComment.WithContext(ctx).
		Where(r.q.EpisodeComment.FieldID.Eq(episodeID)).
		Offset(offset).Limit(limit).
		Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}
	result := make([]model.EpisodeComment, len(s))
	for _, v := range s {
		result = append(result, conventDao2Post(v))
	}
	return result, nil
}

func (r mysqlRepo) AddNewComment(ctx context.Context, comment model.EpisodeComment) error {
	s, err := r.q.WithContext(ctx).EpisodeComment.
		Order(r.q.EpisodeComment.PostID).Last()
	if err != nil {
		return errgo.Wrap(err, "dal")
	}
	id := s.PostID + 1

	err = r.q.WithContext(ctx).EpisodeComment.
		Create(&dao.EpisodeComment{
			PostID:           id,
			FieldID:          comment.Field,
			UserID:           comment.User,
			RelatedMessageID: comment.Related,
			CreatedTime:      int32(comment.CreatedAt.Unix()),
			Content:          comment.Content,
		})
	if err != nil {
		return errgo.Wrap(err, "dal")
	}
	return nil
}

func (r mysqlRepo) DeleteComment(ctx context.Context, episodeID model.EpisodeID, userId model.UserID, commentID model.CommentID) error {
	res, err := r.q.WithContext(ctx).EpisodeComment.
		Where(r.q.EpisodeComment.PostID.Eq(commentID)).Delete()
	if err != nil {
		return errgo.Wrap(err, "dal")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func conventDao2Post(dao *dao.EpisodeComment) model.EpisodeComment {
	return model.EpisodeComment{
		ID:        dao.PostID,
		Field:     dao.FieldID,
		User:      dao.UserID,
		Related:   dao.RelatedMessageID,
		CreatedAt: time.Unix(int64(dao.CreatedTime), 0),
		Content:   dao.Content,
	}
}
