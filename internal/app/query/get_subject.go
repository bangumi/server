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
)

func (q Query) GetSubject(ctx context.Context, user domain.Auth, subjectID model.SubjectID) (model.Subject, error) {
	subject, err := q.getSubject(ctx, subjectID)
	if err != nil {
		return model.Subject{}, err
	}

	if !auth.AllowSubject(user, subject) {
		return model.Subject{}, domain.ErrSubjectNotFound
	}

	return subject, nil
}

func (q Query) getSubject(ctx context.Context, id model.SubjectID) (model.Subject, error) {
	var key = cachekey.Subject(id)

	// try to read from cache
	var r model.Subject
	ok, err := q.cache.Get(ctx, key, &r)
	if err != nil {
		return r, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		return r, nil
	}

	r, err = q.subject.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return r, domain.ErrSubjectNotFound
		}

		return r, errgo.Wrap(err, "SubjectService.Get")
	}

	if e := q.cache.Set(ctx, key, r, time.Minute); e != nil {
		q.log.Error("can't set response to cache", zap.Error(e))
	}

	return r, nil
}
