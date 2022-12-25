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
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic"
	"github.com/bangumi/server/internal/pkg/generic/set"
	"github.com/bangumi/server/internal/pkg/generic/slice"
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
	u domain.Auth,
	subjectID model.SubjectID,
	req UpdateCollectionRequest,
) error {
	ctl.log.Info("try to update collection", subjectID.Zap(), u.ID.Zap(), zap.Reflect("'req", req))

	collect, err := ctl.collection.GetSubjectCollection(ctx, u.ID, subjectID)
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

	if comment := req.Comment.Default(collect.Comment); comment != "" {
		if ctl.dam.NeedReview(comment) {
			privacy = null.New(model.CollectPrivacyBan)
		}
	}

	if slice.Any(req.Tags, ctl.dam.NeedReview) {
		privacy = null.New(model.CollectPrivacyBan)
	}

	txErr := ctl.tx.Transaction(func(tx *query.Query) error {
		err := ctl.collection.WithQuery(tx).UpdateSubjectCollection(ctx, u.ID, subjectID, domain.SubjectCollectionUpdate{
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
			ctl.log.Error("failed to update user collection info", zap.Error(err))
			return errgo.Wrap(err, "collectionRepo.UpdateSubjectCollection")
		}
		err = ctl.saveTimeLineSubject(ctx, u, subjectID, req, tx)
		if err != nil {
			ctl.log.Error("failed to create associated timeline", zap.Error(err))
			return errgo.Wrap(err, "timelineRepo.Create")
		}
		return nil
	})

	return txErr
}

func (ctl Ctrl) saveTimeLineSubject(
	ctx context.Context,
	u domain.Auth,
	subjectID model.SubjectID,
	req UpdateCollectionRequest,
	tx *query.Query,
) error {
	if !(req.VolStatus.Set || req.EpStatus.Set) {
		return nil
	}

	sj, err := ctl.subject.Get(ctx, subjectID, subject.Filter{NSFW: null.Bool{Value: true, Set: !u.AllowNSFW()}})
	if err != nil {
		return errgo.Wrap(err, "subject.Get")
	}
	err = ctl.timeline.WithQuery(tx).Create(ctx, makeTimeLineSubject(req, sj))
	return errgo.Wrap(err, "timeline.Create")
}

func makeTimeLineSubject(req UpdateCollectionRequest, sj model.Subject) *model.TimeLine {
	sidStr := strconv.Itoa(int(sj.ID))

	tlMeta := &model.TimeLineMeta{
		UID:      req.UID,
		Related:  sidStr,
		Dateline: uint32(time.Now().Unix()),
	}

	tlMemo := &model.TimeLineMemo{
		Cat:  model.TimeLineCatSubject,
		Type: convSubjectType(req, sj),
		Content: &model.TimeLineMemoContent{
			TimeLineSubjectMemo: &model.TimeLineSubjectMemo{
				ID:             sidStr,
				TypeID:         string(req.Type.Default(0)),
				Name:           sj.Name,
				NameCN:         sj.NameCN,
				Series:         strconv.Itoa(generic.BtoI(sj.Series)),
				CollectComment: req.Comment.Default(""),
				CollectRate:    int(req.Rate.Default(0)),
			},
		},
	}

	tlImg := model.TimeLineImage{
		SubjectID: &sidStr,
		Images:    &sj.Image,
	}

	return &model.TimeLine{
		TimeLineMeta:   tlMeta,
		TimeLineMemo:   tlMemo,
		TimeLineImages: model.TimeLineImages{tlImg},
	}
}

func convSubjectType(req UpdateCollectionRequest, sj model.Subject) uint16 {
	original := req.Type.Default(0)
	st := sj.TypeID
	l, ok := subjectTypeMap()[st]
	if !ok {
		return uint16(original)
	}
	if original < 1 || original > 5 {
		return uint16(original)
	}
	return l[original]
}

func subjectTypeMap() map[model.SubjectType][]uint16 {
	return map[model.SubjectType][]uint16{
		model.SubjectTypeBook:  {0, 1, 5, 9, 13, 14},
		model.SubjectTypeAnime: {0, 2, 6, 10, 13, 14},
		model.SubjectTypeMusic: {0, 3, 7, 11, 13, 14},
		model.SubjectTypeGame:  {0, 4, 8, 12, 13, 14},
		model.SubjectTypeReal:  {0, 2, 6, 10, 13, 14},
	}
}

func (ctl Ctrl) UpdateEpisodesCollection(
	ctx context.Context,
	u domain.Auth,
	subjectID model.SubjectID,
	episodeIDs []model.EpisodeID,
	t model.EpisodeCollection,
) error {
	if _, err := ctl.GetSubject(ctx, u, subjectID); err != nil {
		return err
	}

	ctl.log.Info("try to update collection info", subjectID.Zap(),
		u.ID.Zap(), zap.Reflect("episode_ids", episodeIDs))

	episodes, err := ctl.episode.List(ctx, subjectID, domain.EpisodeFilter{}, 0, 0)
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

	return ctl.tx.Transaction(
		ctl.updateEpisodesCollectionTx(ctx, u, subjectID, episodeIDs, t, time.Now()),
	)
}

func (ctl Ctrl) UpdateEpisodeCollection(
	ctx context.Context,
	u domain.Auth,
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
	u domain.Auth,
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

		err = collectionTx.UpdateSubjectCollection(ctx, u.ID, subjectID, domain.SubjectCollectionUpdate{
			EpStatus: null.New(uint32(epStatus)),
		}, at)
		if err != nil {
			ctl.log.Error("failed to update user collection info", zap.Error(err))
			return errgo.Wrap(err, "collectionRepo.UpdateSubjectCollection")
		}

		return nil
	}
}
