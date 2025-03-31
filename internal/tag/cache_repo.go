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

package tag

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/cache"
)

const cacheTTL = time.Hour * 24

func NewCachedRepo(c cache.RedisCache, r Repo, log *zap.Logger) CachedRepo {
	return cacheRepo{cache: c, repo: r, log: log.Named("subject.CachedRepo")}
}

var _ CachedRepo = cacheRepo{}

type cacheRepo struct {
	cache cache.RedisCache
	repo  Repo
	log   *zap.Logger
}

type cachedTags struct {
	ID   model.SubjectID
	Tags []Tag
}

//nolint:gochecknoglobals
var CachedCount = prometheus.NewCounter(prometheus.CounterOpts{
	Subsystem:   "chii",
	Name:        "query_cached_count_total",
	Help:        "cached sql query count total",
	ConstLabels: map[string]string{"repo": "meta_tags"},
})

//nolint:gochecknoglobals
var TotalCount = prometheus.NewCounter(prometheus.CounterOpts{
	Subsystem:   "chii",
	Name:        "query_count_total",
	Help:        "sql query count total",
	ConstLabels: map[string]string{"repo": "meta_tags"},
})

//nolint:gochecknoinits
func init() {
	prometheus.MustRegister(CachedCount)
	prometheus.MustRegister(TotalCount)
}

// also need to change version in [cachekey.SubjectMetaTag] if schema is changed.

func (r cacheRepo) Get(ctx context.Context, id model.SubjectID, typeID model.SubjectType) ([]Tag, error) {
	TotalCount.Add(1)
	var key = cachekey.SubjectMetaTag(id)

	var s cachedTags
	ok, err := r.cache.Get(ctx, key, &s)
	if err != nil {
		return s.Tags, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		CachedCount.Add(1)
		return s.Tags, nil
	}

	tags, err := r.repo.Get(ctx, id, typeID)
	if err != nil {
		return tags, err
	}

	if e := r.cache.Set(ctx, key, cachedTags{ID: id, Tags: tags}, cacheTTL); e != nil {
		r.log.Error("can't set response to cache", zap.Error(e))
	}

	return tags, nil
}

func (r cacheRepo) GetByIDs(ctx context.Context, ids []model.SubjectID) (map[model.SubjectID][]Tag, error) {
	result := make(map[model.SubjectID][]Tag, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	TotalCount.Add(float64(len(ids)))

	var tags []cachedTags

	err := r.cache.MGet(ctx, lo.Map(ids, func(item model.SubjectID, index int) string {
		return cachekey.SubjectMetaTag(item)
	}), &tags)
	if err != nil {
		return nil, errgo.Wrap(err, "cache.MGet")
	}

	CachedCount.Add(float64(len(tags)))
	for _, tag := range tags {
		result[tag.ID] = tag.Tags
	}

	var missing = make([]model.SubjectID, 0, len(ids))
	for _, id := range ids {
		if _, ok := result[id]; !ok {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return result, nil
	}

	missingFromCache, err := r.repo.GetByIDs(ctx, missing)
	if err != nil {
		return nil, err
	}
	for id, tag := range missingFromCache {
		result[id] = tag
		err = r.cache.Set(ctx, cachekey.SubjectMetaTag(id), cachedTags{
			ID:   id,
			Tags: tag,
		}, cacheTTL)
		if err != nil {
			return nil, errgo.Wrap(err, "cache.Set")
		}
	}

	return result, nil
}
