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
	"time"

	"github.com/samber/lo"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
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
	Type      null.Null[collection.SubjectCollection]
	Rate      null.Uint8
	Private   null.Bool
}

func (ctl Ctrl) UpdateSubjectCollection(
	ctx context.Context,
	u auth.Auth,
	subject model.Subject,
	req UpdateCollectionRequest,
	allowCreate bool,
) error {
	met := ctl.collection.UpdateSubjectCollection
	if allowCreate {
		met = ctl.collection.UpdateOrCreateSubjectCollection
	}
	err := met(ctx, u.ID, subject, time.Now(), req.IP,
		func(ctx context.Context, s *collection.Subject) (*collection.Subject, error) {
			if req.Comment.Set {
				s.ShadowBan(ctl.dam.NeedReview(req.Comment.Value))
				if e := s.UpdateComment(req.Comment.Value); e != nil {
					return nil, e
				}
			}

			if req.Tags != nil {
				s.ShadowBan(lo.ContainsBy(req.Tags, ctl.dam.NeedReview))
				if e := s.UpdateTags(req.Tags); e != nil {
					return nil, e
				}
			}

			if req.Private.Set {
				if req.Private.Value {
					s.MakePrivate()
				} else {
					s.MakePublic()
				}
			}

			if req.VolStatus.Set {
				s.UpdateVols(req.VolStatus.Value)
			}

			if req.EpStatus.Set {
				s.UpdateEps(req.EpStatus.Value)
			}

			if req.Type.Set {
				s.UpdateType(req.Type.Value)
			}

			if req.Rate.Set {
				if e := s.UpdateRate(req.Rate.Value); e != nil {
					return nil, e
				}
			}
			return s, nil
		},
	)
	if err != nil {
		return err
	}

	return ctl.mayCreateTimeline(ctx, u, req, subject.ID)
}

func (ctl Ctrl) mayCreateTimeline(
	ctx context.Context,
	u auth.Auth,
	req UpdateCollectionRequest,
	subjectID model.SubjectID,
) error {
	collect, err := ctl.collection.GetSubjectCollection(ctx, u.ID, subjectID)
	if err != nil {
		if errors.Is(err, gerr.ErrSubjectNotCollected) {
			ctl.log.Error("failed to create associated timeline, can't get collection ID",
				zap.Error(err), zap.Uint32("user_id", u.ID), zap.Uint32("subject_id", subjectID))
			return nil
		}
		return err
	}

	if collect.Private {
		return nil
	}

	if req.Type.Set {
		sj, err := ctl.subjectCached.Get(ctx, subjectID, subject.Filter{})
		if err != nil {
			return err
		}

		err = ctl.timeline.ChangeSubjectCollection(ctx,
			u.ID, sj, req.Type.Value, collect.ID, req.Comment.Value, req.Rate.Value)
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
		err = ctl.timeline.ChangeSubjectProgress(ctx, u.ID, sj, req.EpStatus.Value, req.VolStatus.Value)
		if err != nil {
			ctl.log.Error("failed to create associated timeline", zap.Error(err))
			return errgo.Wrap(err, "timelineRepo.Create")
		}
	}

	return nil
}
