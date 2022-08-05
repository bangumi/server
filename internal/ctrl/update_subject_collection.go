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
	"fmt"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/set"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/pkg/null"
)

type UpdateCollectionRequest struct {
	IP string

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
	u domain.Auth,
	subjectID model.SubjectID,
	req UpdateCollectionRequest,
) error {
	ctl.log.Info("try to update collection", log.SubjectID(subjectID), log.UserID(u.ID), zap.Reflect("'req", req))

	var privacy null.Null[model.CollectPrivacy]
	if req.Private.Set {
		if req.Private.Value {
			privacy = null.New(model.CollectPrivacyNone)
		} else {
			privacy = null.New(model.CollectPrivacySelf)
		}
	}

	if req.Comment.HasValue() {
		if ctl.dam.NeedReview(req.Comment.Value) {
			privacy = null.New(model.CollectPrivacyBan)
		}
	}

	if slice.Any(req.Tags, ctl.dam.NeedReview) {
		privacy = null.New(model.CollectPrivacyBan)
	}

	err := ctl.tx.Transaction(func(tx *query.Query) error {
		err := ctl.collection.WithQuery(tx).UpdateSubjectCollection(ctx, u.ID, subjectID, domain.SubjectCollectionUpdate{
			Comment:   req.Comment,
			Tags:      req.Tags,
			VolStatus: req.VolStatus,
			EpStatus:  req.EpStatus,
			Type:      req.Type,
			Rate:      req.Rate,
			Privacy:   privacy,
		})
		if err != nil {
			ctl.log.Error("failed to update user collection info", zap.Error(err))
			return errgo.Wrap(err, "collectionRepo.UpdateSubjectCollection")
		}

		return nil
	})

	return err
}

func (ctl Ctrl) UpdateEpisodeCollection(
	ctx context.Context,
	u domain.Auth,
	subjectID model.SubjectID,
	episodeIDs []model.EpisodeID,
	t model.EpisodeCollection,
) error {
	if _, err := ctl.GetSubject(ctx, u, subjectID); err != nil {
		return err
	}

	ctl.log.Info("try to update collection info", log.SubjectID(subjectID),
		log.UserID(u.ID), zap.Reflect("episode_ids", episodeIDs))

	return ctl.tx.Transaction(ctl.updateEpisodeCollectionTx(ctx, u, subjectID, episodeIDs, t))
}

func (ctl Ctrl) updateEpisodeCollectionTx(
	ctx context.Context,
	u domain.Auth,
	subjectID model.SubjectID,
	episodeIDs []model.EpisodeID,
	t model.EpisodeCollection,
) func(tx *query.Query) error {
	return func(tx *query.Query) error {
		episodeTx := ctl.episode.WithQuery(tx)
		collectionTx := ctl.collection.WithQuery(tx)

		episodeCount, err := episodeTx.Count(ctx, subjectID)
		if err != nil {
			return errgo.Wrap(err, "episodeRepo.Count")
		}

		episodes, err := episodeTx.List(ctx, subjectID, int(episodeCount), 0)
		if err != nil {
			return errgo.Wrap(err, "episodeRepo.List")
		}

		eIDs := set.FromSlice(slice.Map(episodes, func(e model.Episode) model.EpisodeID {
			return e.ID
		}))

		for _, d := range episodeIDs {
			if !eIDs.Has(d) {
				return fmt.Errorf("%w: episode %d is not episodes of subject %d", ErrInvalidInput, d, subjectID)
			}
		}

		ec, err := collectionTx.UpdateEpisodeCollection(ctx, u.ID, subjectID, episodeIDs, t)
		if err != nil {
			return errgo.Wrap(err, "UpdateEpisodeCollection")
		}

		epStatus := len(ec)

		err = collectionTx.UpdateSubjectCollection(ctx, u.ID, subjectID, domain.SubjectCollectionUpdate{
			EpStatus: null.NewUint32(uint32(epStatus)),
		})
		if err != nil {
			ctl.log.Error("failed to update user collection info", zap.Error(err))
			return errgo.Wrap(err, "collectionRepo.UpdateSubjectCollection")
		}

		return nil
	}
}
