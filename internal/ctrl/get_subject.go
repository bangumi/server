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

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/ctrl/internal/cachekey"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/gmap"
	"github.com/bangumi/server/internal/pkg/generic/set"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
)

func (ctl Ctrl) GetSubject(ctx context.Context, user domain.Auth, subjectID model.SubjectID) (model.Subject, error) {
	s, err := ctl.getSubject(ctx, subjectID)
	if err != nil {
		return model.Subject{}, err
	}

	if !auth.AllowReadSubject(user, s) {
		return model.Subject{}, domain.ErrSubjectNotFound
	}

	return s, nil
}

func (ctl Ctrl) GetSubjectNoRedirect(
	ctx context.Context,
	user domain.Auth,
	subjectID model.SubjectID,
) (model.Subject, error) {
	s, err := ctl.getSubject(ctx, subjectID)
	if err != nil {
		return model.Subject{}, err
	}

	if s.Redirect != 0 {
		return model.Subject{}, domain.ErrSubjectNotFound
	}

	if !auth.AllowReadSubject(user, s) {
		return model.Subject{}, domain.ErrSubjectNotFound
	}

	return s, nil
}

type SubjectFilter struct {
	NSFW null.Bool
}

func (ctl Ctrl) GetSubjectByIDs(
	ctx context.Context,
	subjectIDs []model.SubjectID,
	filter SubjectFilter,
) (map[model.SubjectID]model.Subject, error) {
	b := ctl.cache.GetMany(ctx, slice.Map(subjectIDs, cachekey.Subject))
	result, err := cache.UnmarshalMany(b, model.Subject.GetID)
	if err != nil {
		return nil, errgo.Wrap(err, "cache.GetMany")
	}

	for key, item := range result {
		if filter.NSFW.Set && (filter.NSFW.Value != item.NSFW) {
			delete(result, key)
		}
	}

	var notCachedID = set.FromSlice(subjectIDs).Removes(gmap.Keys(result)...).ToSlice()
	notCachedSubjects, err := ctl.subject.GetByIDs(ctx, notCachedID, subject.Filter{NSFW: filter.NSFW})
	if err != nil {
		return nil, errgo.Wrap(err, "failed to get subjects")
	}

	if err := ctl.cache.SetMany(ctx, cache.MarshalMany(notCachedSubjects, cachekey.Subject), time.Minute); err != nil {
		ctl.log.Error("cache.SetMany", zap.Error(err))
	}

	return gmap.Merge(result, notCachedSubjects), nil
}

func (ctl Ctrl) getSubject(ctx context.Context, id model.SubjectID) (model.Subject, error) {
	ctl.metricSubjectQueryCount.Inc(1)
	var key = cachekey.Subject(id)

	// try to read from cache
	var r model.Subject
	ok, err := ctl.cache.Get(ctx, key, &r)
	if err != nil {
		return r, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		ctl.metricSubjectQueryCached.Inc(1)
		return r, nil
	}

	r, err = ctl.subject.Get(ctx, id, subject.Filter{})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return r, domain.ErrSubjectNotFound
		}

		return r, errgo.Wrap(err, "SubjectRepo.Get")
	}

	if e := ctl.cache.Set(ctx, key, r, time.Minute); e != nil {
		ctl.log.Error("can't set response to cache", zap.Error(e))
	}

	return r, nil
}
