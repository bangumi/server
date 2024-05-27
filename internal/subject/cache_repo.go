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

package subject

import (
	"context"
	"time"

	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/cache"
)

func NewCachedRepo(c cache.RedisCache, r Repo, log *zap.Logger) CachedRepo {
	return cacheRepo{cache: c, repo: r, log: log.Named("subject.CachedRepo")}
}

var _ CachedRepo = cacheRepo{}

type cacheRepo struct {
	cache cache.RedisCache
	repo  Repo
	log   *zap.Logger
}

func (r cacheRepo) Get(ctx context.Context, id model.SubjectID, filter Filter) (model.Subject, error) {
	var key = cachekey.Subject(id)

	// try to read from cache
	var s model.Subject
	ok, err := r.cache.Get(ctx, key, &s)
	if err != nil {
		return s, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		return s, nil
	}

	s, err = r.repo.Get(ctx, id, filter)
	if err != nil {
		return s, err
	}

	if e := r.cache.Set(ctx, key, s, time.Minute); e != nil {
		r.log.Error("can't set response to cache", zap.Error(e))
	}

	return s, nil
}

func (r cacheRepo) GetByIDs(
	ctx context.Context, ids []model.SubjectID, filter Filter,
) (map[model.SubjectID]model.Subject, error) {
	return r.repo.GetByIDs(ctx, ids, filter)
}

func (r cacheRepo) Count(ctx context.Context, filter BrowseFilter) (int64, error) {
	hash, err := filter.Hash()
	if err != nil {
		return 0, err
	}
	key := cachekey.SubjectBrowseCount(hash)

	var s int64
	ok, err := r.cache.Get(ctx, key, &s)
	if err != nil {
		return s, errgo.Wrap(err, "cache.Get")
	}
	if ok {
		return s, nil
	}

	s, err = r.repo.Count(ctx, filter)
	if err != nil {
		return s, err
	}
	if e := r.cache.Set(ctx, key, s, 10*time.Minute); e != nil {
		r.log.Error("can't set response to cache", zap.Error(e))
	}

	return s, nil
}

func (r cacheRepo) Browse(
	ctx context.Context, filter BrowseFilter, limit, offset int,
) ([]model.Subject, error) {
	hash, err := filter.Hash()
	if err != nil {
		return nil, err
	}
	key := cachekey.SubjectBrowse(hash)

	var subjects []model.Subject
	ok, err := r.cache.Get(ctx, key, &subjects)
	if err != nil {
		return nil, errgo.Wrap(err, "cache.Get")
	}
	if ok {
		return subjects, nil
	}

	subjects, err = r.repo.Browse(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}
	ttl := 24 * time.Hour
	if offset > 0 {
		ttl = 10 * time.Minute
	}
	if e := r.cache.Set(ctx, key, subjects, ttl); e != nil {
		r.log.Error("can't set response to cache", zap.Error(e))
	}

	return subjects, nil

}

func (r cacheRepo) GetPersonRelated(
	ctx context.Context, personID model.PersonID,
) ([]domain.SubjectPersonRelation, error) {
	return r.repo.GetPersonRelated(ctx, personID)
}

func (r cacheRepo) GetCharacterRelated(
	ctx context.Context, characterID model.CharacterID,
) ([]domain.SubjectCharacterRelation, error) {
	return r.repo.GetCharacterRelated(ctx, characterID)
}

func (r cacheRepo) GetSubjectRelated(
	ctx context.Context, subjectID model.SubjectID,
) ([]domain.SubjectInternalRelation, error) {
	return r.repo.GetSubjectRelated(ctx, subjectID)
}

func (r cacheRepo) GetActors(
	ctx context.Context, subjectID model.SubjectID, characterIDs []model.CharacterID,
) (map[model.CharacterID][]model.PersonID, error) {
	return r.repo.GetActors(ctx, subjectID, characterIDs)
}
