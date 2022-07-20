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

package query

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/app/internal/cachekey"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/gmap"
)

func (q Query) GetSubject(ctx context.Context, user domain.Auth, subjectID model.SubjectID) (model.Subject, error) {
	subject, err := q.getSubject(ctx, subjectID)
	if err != nil {
		return model.Subject{}, err
	}

	if !auth.AllowReadSubject(user, subject) {
		return model.Subject{}, domain.ErrSubjectNotFound
	}

	return subject, nil
}

func (q Query) GetSubjectNoRedirect(
	ctx context.Context,
	user domain.Auth,
	subjectID model.SubjectID,
) (model.Subject, error) {
	subject, err := q.getSubject(ctx, subjectID)
	if err != nil {
		return model.Subject{}, err
	}

	if subject.Redirect != 0 {
		return model.Subject{}, domain.ErrSubjectNotFound
	}

	if !auth.AllowReadSubject(user, subject) {
		return model.Subject{}, domain.ErrSubjectNotFound
	}

	return subject, nil
}

func (q Query) GetSubjectByIDs(
	ctx context.Context,
	subjectIDs ...model.SubjectID,
) (map[model.SubjectID]model.Subject, error) {
	q.metricSubjectQueryCount.Inc(int64(len(subjectIDs)))
	var notCached = make([]model.SubjectID, 0, len(subjectIDs))

	var result = make(map[model.SubjectID]model.Subject, len(subjectIDs))
	for _, subjectID := range subjectIDs {
		key := cachekey.Subject(subjectID)
		var s model.Subject
		ok, err := q.cache.Get(ctx, key, &s)
		if err != nil {
			return nil, errgo.Wrap(err, "cache.Get")
		}

		if ok {
			q.metricSubjectQueryCached.Inc(1)
			result[subjectID] = s
		} else {
			notCached = append(notCached, subjectID)
		}
	}

	newSubjectMap, err := q.subject.GetByIDs(ctx, notCached...)
	if err != nil {
		return nil, errgo.Wrap(err, "failed to get subjects")
	}

	for subjectID, subject := range newSubjectMap {
		err = q.cache.Set(ctx, cachekey.Subject(subjectID), subject, time.Minute)
		if err != nil {
			q.log.Error("failed to set subject cache")
		}
	}

	gmap.Copy(result, newSubjectMap)

	return result, nil
}

func (q Query) getSubject(ctx context.Context, id model.SubjectID) (model.Subject, error) {
	q.metricSubjectQueryCount.Inc(1)
	var key = cachekey.Subject(id)

	// try to read from cache
	var r model.Subject
	ok, err := q.cache.Get(ctx, key, &r)
	if err != nil {
		return r, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		q.metricSubjectQueryCached.Inc(1)
		return r, nil
	}

	r, err = q.subject.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return r, domain.ErrSubjectNotFound
		}

		return r, errgo.Wrap(err, "SubjectRepo.Get")
	}

	if e := q.cache.Set(ctx, key, r, time.Minute); e != nil {
		q.log.Error("can't set response to cache", zap.Error(e))
	}

	return r, nil
}
