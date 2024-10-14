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
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/meilisearch/meilisearch-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/trim21/errgo"
	"github.com/trim21/pkg/queue"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/subject"
)

// New provide a search app is AppConfig.MeiliSearchURL is empty string, return nope search client.
//
// see `MeiliSearchURL` and `MeiliSearchKey` in [config.AppConfig].
func New(
	cfg config.AppConfig,
	subjectRepo subject.Repo,
	log *zap.Logger,
	query *query.Query,
) (Client, error) {
	if cfg.Search.MeiliSearch.URL == "" {
		return NoopClient{}, nil
	}

	if subjectRepo == nil {
		panic("nil SubjectRepo")
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

	c := &client{
		meili:        meili,
		q:            query,
		subject:      "subjects",
		subjectIndex: meili.Index("subjects"),
		log:          log.Named("search"),
		subjectRepo:  subjectRepo,
	}

	if cfg.AppType != config.AppTypeCanal {
		return c, nil
	}

	return c, c.canalInit(cfg)
}

func (c *client) canalInit(cfg config.AppConfig) error {
	if cfg.Search.SearchBatchSize <= 0 {
		// nolint: goerr113
		return fmt.Errorf("config.SearchBatchSize should >= 0, current %d", cfg.Search.SearchBatchSize)
	}

	if cfg.Search.SearchBatchInterval <= 0 {
		// nolint: goerr113
		return fmt.Errorf("config.SearchBatchInterval should >= 0, current %d", cfg.Search.SearchBatchInterval)
	}

	c.queue = queue.NewBatchedDedupe[subjectIndex](
		c.sendBatch,
		cfg.Search.SearchBatchSize,
		cfg.Search.SearchBatchInterval,
		func(items []subjectIndex) []subjectIndex {
			// lo.UniqBy 会保留第一次出现的元素，reverse 之后会保留新的数据
			return lo.UniqBy(lo.Reverse(items), func(item subjectIndex) model.SubjectID {
				return item.ID
			})
		},
	)

	prometheus.DefaultRegisterer.MustRegister(
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: "chii",
				Name:      "meilisearch_queue_batch",
				Help:      "meilisearch update queue batch size",
			},
			func() float64 {
				return float64(c.queue.Len())
			},
		))

	shouldCreateIndex, err := c.needFirstRun()
	if err != nil {
		return err
	}

	if shouldCreateIndex {
		go c.firstRun()
	}

	return nil
}

type client struct {
	subjectRepo  subject.Repo
	meili        meilisearch.ServiceManager
	q            *query.Query
	subjectIndex meilisearch.IndexManager
	log          *zap.Logger
	subject      string
	queue        *queue.Batched[subjectIndex]
}

func (c *client) Close() {
	if c.queue != nil {
		c.queue.Close()
	}
}

// OnSubjectUpdate is the hook called by canal.
func (c *client) OnSubjectUpdate(ctx context.Context, id model.SubjectID) error {
	s, err := c.subjectRepo.Get(ctx, id, subject.Filter{})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return nil
		}
		return errgo.Wrap(err, "subjectRepo.Get")
	}

	if s.Redirect != 0 || s.Ban != 0 {
		return c.DeleteSubject(ctx, id)
	}

	extracted := extractSubject(&s)

	c.queue.Push(extracted)

	return nil
}

// OnSubjectDelete is the hook called by canal.
func (c *client) OnSubjectDelete(_ context.Context, id model.SubjectID) error {
	_, err := c.subjectIndex.DeleteDocument(strconv.FormatUint(uint64(id), 10))

	return errgo.Wrap(err, "search")
}

// UpsertSubject add subject to search backend.
func (c *client) sendBatch(items []subjectIndex) {
	c.log.Debug("send batch to meilisearch", zap.Int("len", len(items)))
	err := retry.Do(
		func() error {
			_, err := c.subjectIndex.UpdateDocuments(items, "id")
			return err
		},
		retry.OnRetry(func(n uint, err error) {
			c.log.Warn("failed to send batch", zap.Uint("attempt", n), zap.Error(err))
		}),

		retry.DelayType(retry.BackOffDelay),
		retry.Delay(time.Microsecond*100),
		retry.Attempts(5), //nolint:mnd
		retry.RetryIf(func(err error) bool {
			var r = &meilisearch.Error{}
			return errors.As(err, &r)
		}),
	)

	if err != nil {
		c.log.Error("failed to send batch", zap.Error(err))
	}
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

//nolint:funlen
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

	c.log.Info("set sortable attributes", zap.Strings("attributes", *getAttributes("sortable")))
	_, err = subjectIndex.UpdateSortableAttributes(getAttributes("sortable"))
	if err != nil {
		c.log.Fatal("failed to update search index sortable attributes", zap.Error(err))
		return
	}

	c.log.Info("set filterable attributes", zap.Strings("attributes", *getAttributes("filterable")))
	_, err = subjectIndex.UpdateFilterableAttributes(getAttributes("filterable"))
	if err != nil {
		c.log.Fatal("failed to update search index filterable attributes", zap.Error(err))
		return
	}

	c.log.Info("set searchable attributes", zap.Strings("attributes", *getAttributes("searchable")))
	_, err = subjectIndex.UpdateSearchableAttributes(getAttributes("searchable"))
	if err != nil {
		c.log.Fatal("failed to update search index searchable attributes", zap.Error(err))
		return
	}

	c.log.Info("set ranking rules", zap.Strings("rule", *rankRule()))
	_, err = subjectIndex.UpdateRankingRules(rankRule())
	if err != nil {
		c.log.Fatal("failed to update search index searchable attributes", zap.Error(err))
		return
	}

	ctx := context.Background()

	maxSubject, err := c.q.Subject.WithContext(ctx).Limit(1).Order(c.q.Subject.ID.Desc()).Take()
	if err != nil {
		c.log.Fatal("failed to get current max subject id", zap.Error(err))
		return
	}

	c.log.Info(fmt.Sprintf("run full search index with max subject id %d", maxSubject.ID))

	width := len(strconv.Itoa(int(maxSubject.ID)))
	for i := model.SubjectID(1); i < maxSubject.ID; i++ {
		if i%10000 == 0 {
			c.log.Info(fmt.Sprintf("progress %*d/%d", width, i, maxSubject.ID))
		}

		err := c.OnSubjectUpdate(ctx, i)
		if err != nil {
			c.log.Error("error when updating subject", zap.Error(err))
		}
	}
}

func getAttributes(tag string) *[]string {
	rt := reflect.TypeOf(subjectIndex{})
	var s []string
	for i := 0; i < rt.NumField(); i++ {
		t, ok := rt.Field(i).Tag.Lookup(tag)
		if !ok {
			continue
		}

		if t != "true" {
			continue
		}

		s = append(s, getJSONFieldName(rt.Field(i)))
	}

	return &s
}

func getJSONFieldName(f reflect.StructField) string {
	t := f.Tag.Get("json")
	if t == "" {
		return f.Name
	}

	return strings.Split(t, ",")[0]
}
