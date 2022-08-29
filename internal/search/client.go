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
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger/log"
)

// New provide a search app is AppConfig.MeiliSearchURL is empty string, return nope search client.
//
// see `MeiliSearchURL` and `MeiliSearchKey` in [config.AppConfig].
func New(
	c config.AppConfig,
	subjectRepo domain.SubjectRepo,
	log *zap.Logger,
	query *query.Query,
) (Client, error) {
	if c.MeiliSearchURL == "" {
		return NoopClient{}, nil
	}

	if subjectRepo == nil {
		panic("nil SubjectRepo")
	}
	if _, err := url.Parse(c.MeiliSearchURL); err != nil {
		return nil, errgo.Wrap(err, "url.Parse")
	}

	meili := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    c.MeiliSearchURL,
		APIKey:  c.MeiliSearchKey,
		Timeout: time.Second,
	})

	if _, err := meili.GetVersion(); err != nil {
		return nil, errgo.Wrap(err, "meilisearch")
	}

	client := &client{
		meili:        meili,
		q:            query,
		subject:      "subjects",
		subjectIndex: meili.Index("subjects"),
		log:          log.Named("search"),
		subjectRepo:  subjectRepo,
	}

	shouldCreateIndex, err := client.needFirstRun()
	if err != nil {
		return nil, err
	}

	if shouldCreateIndex {
		go client.firstRun()
	}

	return client, nil
}

type client struct {
	subjectRepo  domain.SubjectRepo
	meili        *meilisearch.Client
	q            *query.Query
	subjectIndex *meilisearch.Index
	log          *zap.Logger
	subject      string
}

// OnSubjectUpdate is the hook called by canal.
func (c *client) OnSubjectUpdate(ctx context.Context, id model.SubjectID) error {
	s, err := c.subjectRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.DeleteSubject(ctx, id)
		}

		c.log.Error("unexpected error get subject from mysql", zap.Error(err), log.SubjectID(id))
		return errgo.Wrap(err, "subjectRepo.Get")
	}

	extracted := extractSubject(&s)

	return c.upsertSubject(ctx, extracted)
}

// OnSubjectDelete is the hook called by canal.
func (c *client) OnSubjectDelete(_ context.Context, id model.SubjectID) error {
	_, err := c.subjectIndex.DeleteDocument(strconv.FormatUint(uint64(id), 10))

	return errgo.Wrap(err, "search")
}

// UpsertSubject add subject to search backend.
func (c *client) upsertSubject(_ context.Context, s subjectIndex) error {
	_, err := c.subjectIndex.UpdateDocuments(s)

	return errgo.Wrap(err, "search")
}

func (c *client) DeleteSubject(_ context.Context, id model.SubjectID) error {
	_, err := c.subjectIndex.Delete(strconv.FormatUint(uint64(id), 10))

	return errgo.Wrap(err, "delete")
}

func (c *client) needFirstRun() (bool, error) {
	if os.Getenv("CHII_SEARCH_INIT") == "true" {
		return true, nil
	}

	index, err := c.meili.GetIndex("subjects")
	if err != nil {
		var e *meilisearch.Error
		if errors.As(err, &e) {
			return true, nil
		}
		return false, errgo.Wrap(err, "get subjects index")
	}

	stat, err := index.GetStats()
	if err != nil {
		return false, errgo.Wrap(err, "get subjects index stats")
	}

	return stat.NumberOfDocuments == 0, nil
}

func (c *client) firstRun() {
	c.log.Info("search initialize")
	_, err := c.meili.CreateIndex(&meilisearch.IndexConfig{
		Uid:        "subjects",
		PrimaryKey: "id",
	})
	if err != nil {
		c.log.Fatal("failed to create search subject index", zap.Error(err))
		return
	}
	subjectIndex := c.meili.Index("subjects")

	_, err = subjectIndex.UpdateFilterableAttributes(&[]string{
		"type",
		"score",
		"nsfw",
		"rank",
		"date",
		"tag",
	})
	if err != nil {
		c.log.Fatal("failed to update search index filterable attributes", zap.Error(err))
		return
	}

	_, err = subjectIndex.UpdateSearchableAttributes(&[]string{
		"name",
		"name_cn",
		"summary",
	})
	if err != nil {
		c.log.Fatal("failed to update search index searchable attributes", zap.Error(err))
		return
	}

	ctx := context.Background()

	maxSubject, err := c.q.Subject.WithContext(ctx).Limit(1).Order(c.q.Subject.ID.Desc()).First()
	if err != nil {
		c.log.Fatal("failed to get current max subject id", zap.Error(err))
		return
	}

	c.log.Info(fmt.Sprintln("run full search index with max subject id", maxSubject.ID))

	for i := model.SubjectID(1); i < maxSubject.ID; i++ {
		err := c.OnSubjectUpdate(ctx, i)
		if err != nil {
			c.log.Error("error when updating subject", zap.Error(err))
		}
	}
}
