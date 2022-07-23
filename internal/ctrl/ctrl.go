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
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/domain"
)

func New(
	episode domain.EpisodeRepo,
	cache cache.Cache,
	subject domain.SubjectRepo,
	person domain.PersonRepo,
	character domain.CharacterRepo,
	collection domain.CollectionRepo,
	metric tally.Scope,
	user domain.UserRepo,
	topic domain.TopicRepo,
	log *zap.Logger,
) Ctrl {
	return Ctrl{
		log:   log.Named("app.query"),
		cache: cache,

		user:       user,
		topic:      topic,
		person:     person,
		episode:    episode,
		subject:    subject,
		collection: collection,
		character:  character,

		metricUserQueryCount:  metric.Counter("app_user_query_count"),
		metricUserQueryCached: metric.Counter("app_user_query_cached_count"),

		metricSubjectQueryCount:  metric.Counter("app_subject_query_count"),
		metricSubjectQueryCached: metric.Counter("app_subject_query_cached_count"),

		metricsEpisodeQueryCount:  metric.Counter("app_episode_query_count"),
		metricsEpisodeQueryCached: metric.Counter("app_subject_query_cached_count"),
	}
}

type Ctrl struct {
	log   *zap.Logger
	cache cache.Cache

	user                  domain.UserRepo
	topic                 domain.TopicRepo
	person                domain.PersonRepo
	episode               domain.EpisodeRepo
	subject               domain.SubjectRepo
	character             domain.CharacterRepo
	collection            domain.CollectionRepo
	metricUserQueryCached tally.Counter
	metricUserQueryCount  tally.Counter

	metricSubjectQueryCached tally.Counter
	metricSubjectQueryCount  tally.Counter

	metricsEpisodeQueryCount  tally.Counter
	metricsEpisodeQueryCached tally.Counter
}
