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
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/samber/lo"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/set"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/subject"
)

func (ctl Ctrl) UpdateEpisodesCollection(
	ctx context.Context,
	u auth.Auth,
	subjectID model.SubjectID,
	episodeIDs []model.EpisodeID,
	t collection.EpisodeCollection,
) error {
	if _, err := ctl.subjectCached.Get(ctx, subjectID, subject.Filter{}); err != nil {
		return err
	}

	ctl.log.Info("try to update collection info", zap.Uint32("subject", subjectID),
		log.User(u.ID), zap.Reflect("episodes", episodeIDs))

	/*
		GORM v1.25.0 起修复了一个 bug，但是被当成 feature 使用了。在该版本之前，Limit 0 认为不是合法的 Limit，会被从 SQL 语句中忽略
		see PR: go-gorm/gorm/pull/6191
		因此这里需要传入 -1 作为 Limit，从而返回全部数据。GORM 会对负数过过滤，不会出现在最终的 SQL 中。
	*/
	episodes, err := ctl.episode.List(ctx, subjectID, episode.Filter{}, -1, 0)
	if err != nil {
		return errgo.Wrap(err, "episodeRepo.List")
	}

	eIDs := set.FromSlice(slice.Map(episodes, episode.Episode.GetID))
	for _, d := range episodeIDs {
		if !eIDs.Has(d) {
			return fmt.Errorf("%w: episode %d is not episodes of subject %d", ErrInvalidInput, d, subjectID)
		}
	}

	err = ctl.tx.Transaction(ctl.updateEpisodesCollectionTx(ctx, u, subjectID, episodeIDs, t, time.Now()))

	if err != nil {
		return err
	}

	episodes = lo.Filter(episodes, func(item episode.Episode, _ int) bool {
		return item.Type == episode.TypeNormal && lo.Contains(episodeIDs, item.ID)
	})

	if len(episodes) == 0 {
		return nil
	}

	sort.Slice(episodes, func(i, j int) bool {
		return !episodes[i].Less(episodes[j])
	})

	e := episodes[0]

	s, err := ctl.subjectCached.Get(ctx, e.SubjectID, subject.Filter{})
	if err != nil {
		return err
	}

	err = ctl.timeline.ChangeEpisodeStatus(ctx, u, s, e)

	return errgo.Wrap(err, "timeline.ChangeEpisodeStatus")
}

func (ctl Ctrl) UpdateEpisodeCollection(
	ctx context.Context,
	u auth.Auth,
	episodeID model.EpisodeID,
	t collection.EpisodeCollection,
) error {
	ctl.log.Info("try to update episode collection info", log.User(u.ID), zap.Uint32("episode", episodeID))

	e, err := ctl.episode.Get(ctx, episodeID)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return gerr.ErrEpisodeNotFound
		}

		return errgo.Wrap(err, "episode.Get")
	}

	err = ctl.tx.Transaction(
		ctl.updateEpisodesCollectionTx(ctx, u, e.SubjectID, []model.EpisodeID{episodeID}, t, time.Now()),
	)

	if err != nil {
		return err
	}

	s, err := ctl.subjectCached.Get(ctx, e.SubjectID, subject.Filter{})
	if err != nil {
		return err
	}

	err = ctl.timeline.ChangeEpisodeStatus(ctx, u, s, e)

	return errgo.Wrap(err, "timeline.ChangeEpisodeStatus")
}

func (ctl Ctrl) updateEpisodesCollectionTx(
	ctx context.Context,
	u auth.Auth,
	subjectID model.SubjectID,
	episodeIDs []model.EpisodeID,
	t collection.EpisodeCollection,
	at time.Time,
) func(tx *query.Query) error {
	return func(tx *query.Query) error {
		collectionTx := ctl.collection.WithQuery(tx)

		_, err := collectionTx.GetSubjectCollection(ctx, u.ID, subjectID)
		if err != nil {
			if errors.Is(err, gerr.ErrNotFound) {
				return gerr.ErrSubjectNotCollected
			}

			return errgo.Wrap(err, "collection.GetSubjectCollection")
		}

		ec, err := collectionTx.UpdateEpisodeCollection(ctx, u.ID, subjectID, episodeIDs, t, at)
		if err != nil {
			return errgo.Wrap(err, "UpdateEpisodeCollection")
		}

		epStatus := len(ec)

		err = collectionTx.UpdateSubjectCollection(ctx, u.ID, subjectID, time.Now(), "",
			func(ctx context.Context, s *collection.Subject) (*collection.Subject, error) {
				s.UpdateEps(uint32(epStatus))
				return s, nil
			})
		if err != nil {
			ctl.log.Error("failed to update user collection info", zap.Error(err))
			return errgo.Wrap(err, "collectionRepo.UpdateSubjectCollection")
		}

		return nil
	}
}
