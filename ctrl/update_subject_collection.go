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
	"github.com/bangumi/server/internal/pkg/null"
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
	ctl.log.Info("try to update collection", subjectID.Zap(), u.ID.Zap(), zap.Reflect("'req", req))

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

	if req.Type.Set {
		sj, err := ctl.GetSubject(ctx, u, subjectID)
		if err != nil {
			return err
		}
		err = ctl.timeline.ChangeSubjectCollection(ctx, u, sj, req.Type.Default(0), req.Comment.Value, req.Rate.Value)
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
	if _, err := ctl.GetSubject(ctx, u, subjectID); err != nil {
		return err
	}

	ctl.log.Info("try to update collection info", subjectID.Zap(),
		u.ID.Zap(), zap.Reflect("episode_ids", episodeIDs))

	episodes, err := ctl.episode.List(ctx, subjectID, episode.Filter{}, 0, 0)
	if err != nil {
		return errgo.Wrap(err, "episodeRepo.List")
	}

	eIDs := set.FromSlice(slice.Map(episodes, func(e episode.Episode) model.EpisodeID {
		return e.ID
	}))

	for _, d := range episodeIDs {
		if !eIDs.Has(d) {
			return fmt.Errorf("%w: episode %d is not episodes of subject %d", ErrInvalidInput, d, subjectID)
		}
	}

	return ctl.tx.Transaction(
		ctl.updateEpisodesCollectionTx(ctx, u, subjectID, episodeIDs, t, time.Now()),
	)
}

func (ctl Ctrl) UpdateEpisodeCollection(
	ctx context.Context,
	u auth.Auth,
	episodeID model.EpisodeID,
	t model.EpisodeCollection,
) error {
	ctl.log.Info("try to update episode collection info", u.ID.Zap(), episodeID.Zap())

	e, err := ctl.episode.Get(ctx, episodeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrEpisodeNotFound
		}

		return errgo.Wrap(err, "episode.Get")
	}

	return ctl.tx.Transaction(
		ctl.updateEpisodesCollectionTx(ctx, u, e.SubjectID, []model.EpisodeID{episodeID}, t, time.Now()),
	)
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
