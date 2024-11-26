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

package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/meilisearch/meilisearch-go"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/person"
	characterSearcher "github.com/bangumi/server/internal/search/character"
	personSearcher "github.com/bangumi/server/internal/search/person"
	"github.com/bangumi/server/internal/search/searcher"
	subjectSearcher "github.com/bangumi/server/internal/search/subject"
	"github.com/bangumi/server/internal/subject"
)

type SearchTarget string

const (
	SearchTargetSubject   SearchTarget = "subject"
	SearchTargetCharacter SearchTarget = "character"
	SearchTargetPerson    SearchTarget = "person"
)

type Client interface {
	Handle(c echo.Context, target SearchTarget) error
	Close()

	EventAdded(ctx context.Context, id uint32, target SearchTarget) error
	EventUpdate(ctx context.Context, id uint32, target SearchTarget) error
	EventDelete(ctx context.Context, id uint32, target SearchTarget) error
}

type Handler interface {
	Handle(c echo.Context, target SearchTarget) error
}

type Search struct {
	searchers map[SearchTarget]searcher.Searcher
}

// New provide a search app is AppConfig.MeiliSearchURL is empty string, return nope search client.
//
// see `MeiliSearchURL` and `MeiliSearchKey` in [config.AppConfig].
func New(
	cfg config.AppConfig,
	subjectRepo subject.Repo,
	characterRepo character.Repo,
	personRepo person.Repo,
	log *zap.Logger,
	query *query.Query,
) (Client, error) {
	if cfg.Search.MeiliSearch.URL == "" {
		return NoopClient{}, nil
	}
	if _, err := url.Parse(cfg.Search.MeiliSearch.URL); err != nil {
		return nil, errgo.Wrap(err, "url.Parse")
	}
	meili := meilisearch.New(
		cfg.Search.MeiliSearch.URL,
		meilisearch.WithAPIKey(cfg.Search.MeiliSearch.Key),
		meilisearch.WithCustomClient(&http.Client{Timeout: cfg.Search.MeiliSearch.Timeout}),
	)
	if _, err := meili.Version(); err != nil {
		return nil, errgo.Wrap(err, "meilisearch")
	}

	subject, err := subjectSearcher.New(cfg, meili, subjectRepo, log, query)
	if err != nil {
		return nil, errgo.Wrap(err, "subject search")
	}
	character, err := characterSearcher.New(cfg, meili, characterRepo, log, query)
	if err != nil {
		return nil, errgo.Wrap(err, "character search")
	}
	person, err := personSearcher.New(cfg, meili, personRepo, log, query)
	if err != nil {
		return nil, errgo.Wrap(err, "person search")
	}

	searchers := map[SearchTarget]searcher.Searcher{
		SearchTargetSubject:   subject,
		SearchTargetCharacter: character,
		SearchTargetPerson:    person,
	}
	s := &Search{
		searchers: searchers,
	}
	return s, nil
}

func (s *Search) Handle(c echo.Context, target SearchTarget) error {
	searcher := s.searchers[target]
	if searcher == nil {
		return fmt.Errorf("searcher not found for %s", target)
	}
	return searcher.Handle(c)
}

func (s *Search) EventAdded(ctx context.Context, id uint32, target SearchTarget) error {
	searcher := s.searchers[target]
	if searcher == nil {
		return fmt.Errorf("searcher not found for %s", target)
	}
	return searcher.OnAdded(ctx, id)
}

func (s *Search) EventUpdate(ctx context.Context, id uint32, target SearchTarget) error {
	searcher := s.searchers[target]
	if searcher == nil {
		return fmt.Errorf("searcher not found for %s", target)
	}
	return searcher.OnUpdate(ctx, id)
}

func (s *Search) EventDelete(ctx context.Context, id uint32, target SearchTarget) error {
	searcher := s.searchers[target]
	if searcher == nil {
		return fmt.Errorf("searcher not found for %s", target)
	}
	return searcher.OnDelete(ctx, id)
}

func (s *Search) Close() {
	for _, searcher := range s.searchers {
		searcher.Close()
	}
}
