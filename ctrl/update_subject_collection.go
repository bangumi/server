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
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/set"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
)

type UpdateCollectionRequest struct {
	IP  string
	UID model.UserID

	Comment   null.String
	Tags      []string
	VolStatus null.Uint32
	EpStatus  null.Uint32
	Type      null.Null[model.SubjectCollection]
	Rate      null.Uint8
	Private   null.Bool
}

func (ctl Ctrl) UpdateCollection(
	ctx context.Context,
	u auth.Auth,
	subjectID model.SubjectID,
	req UpdateCollectionRequest,
) error {
	ctl.log.Info("try to update collection", zap.Uint32("subject_id", subjectID), log.User(u.ID))

	original, err := ctl.collection.GetSubjectCollection(ctx, u.ID, subjectID)
	if err != nil {
		return errgo.Wrap(err, "collectionRepo.GetSubjectCollection")
	}

	var privacy null.Null[model.CollectPrivacy]
	if req.Private.Set {
		if req.Private.Value {
			privacy = null.New(model.CollectPrivacySelf)
		} else {
			privacy = null.New(model.CollectPrivacyNone)
		}
	}

	if comment := req.Comment.Default(original.Comment); comment != "" {
		if ctl.dam.NeedReview(comment) {
			privacy = null.New(model.CollectPrivacyBan)
		}
	}

	if lo.ContainsBy(req.Tags, ctl.dam.NeedReview) {
		privacy = null.New(model.CollectPrivacyBan)
	}

	err = ctl.collection.UpdateSubjectCollection(ctx, u.ID, subjectID, collection.Update{
		IP:        req.IP,
		Comment:   req.Comment,
		Tags:      req.Tags,
		VolStatus: req.VolStatus,
		EpStatus:  req.EpStatus,
		Type:      req.Type,
		Rate:      req.Rate,
		Privacy:   privacy,
	}, time.Now())

	if err != nil {
		return errgo.Wrap(err, "collectionRepo.UpdateSubjectCollection")
	}

	return ctl.mayCreateTimeline(ctx, u, req, subjectID)
}

func (ctl Ctrl) mayCreateTimeline(
	ctx context.Context,
	u auth.Auth,
	req UpdateCollectionRequest,
	subjectID model.SubjectID,
) error {
	if req.Type.Set {
		sj, err := ctl.subjectCached.Get(ctx, subjectID, subject.Filter{})
		if err != nil {
			return err
		}
		err = ctl.timeline.ChangeSubjectCollection(ctx, u, sj, req.Type.Default(0), req.Comment.Value, req.Rate.Value)
		if err != nil {
			ctl.log.Error("failed to create associated timeline", zap.Error(err))
			return errgo.Wrap(err, "timelineRepo.Create")
		}
	}

	if req.EpStatus.Set || req.VolStatus.Set {
		sj, err := ctl.subjectCached.Get(ctx, subjectID, subject.Filter{})
		if err != nil {
			return err
		}
		err = ctl.timeline.ChangeSubjectProgress(ctx, u, sj, req.EpStatus.Value, req.VolStatus.Value)
		if err != nil {
			ctl.log.Error("failed to create associated timeline", zap.Error(err))
			return errgo.Wrap(err, "timelineRepo.Create")
		}
	}

	return nil
}

func (ctl Ctrl) UpdateEpisodesCollection(
	ctx context.Context,
	u auth.Auth,
	subjectID model.SubjectID,
	episodeIDs []model.EpisodeID,
	t model.EpisodeCollection,
) error {
	if _, err := ctl.subjectCached.Get(ctx, subjectID, subject.Filter{}); err != nil {
		return err
	}

	ctl.log.Info("try to update collection info", zap.Uint32("subject", subjectID),
		log.User(u.ID), zap.Reflect("episodes", episodeIDs))

	episodes, err := ctl.episode.List(ctx, subjectID, episode.Filter{}, 0, 0)
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
	t model.EpisodeCollection,
) error {
	ctl.log.Info("try to update episode collection info", log.User(u.ID), zap.Uint32("episode", episodeID))

	e, err := ctl.episode.Get(ctx, episodeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrEpisodeNotFound
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
	t model.EpisodeCollection,
	at time.Time,
) func(tx *query.Query) error {
	return func(tx *query.Query) error {
		collectionTx := ctl.collection.WithQuery(tx)

		_, err := collectionTx.GetSubjectCollection(ctx, u.ID, subjectID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrSubjectNotCollected
			}

			return errgo.Wrap(err, "collection.GetSubjectCollection")
		}

		ec, err := collectionTx.UpdateEpisodeCollection(ctx, u.ID, subjectID, episodeIDs, t, at)
		if err != nil {
			return errgo.Wrap(err, "UpdateEpisodeCollection")
		}

		epStatus := len(ec)

		err = collectionTx.UpdateSubjectCollection(ctx, u.ID, subjectID, collection.Update{
			EpStatus: null.New(uint32(epStatus)),
		}, at)
		if err != nil {
			ctl.log.Error("failed to update user collection info", zap.Error(err))
			return errgo.Wrap(err, "collectionRepo.UpdateSubjectCollection")
		}

		return nil
	}
}
