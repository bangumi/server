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

func New(
	c config.AppConfig,
	subjectRepo domain.SubjectRepo,
	log *zap.Logger,
	query *query.Query,
) (*Client, error) {
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

	var shouldCreateIndex = false
	index, err := meili.GetIndex("subjects")
	if err != nil {
		var e *meilisearch.Error
		if errors.As(err, &e) {
			shouldCreateIndex = true
		} else {
			panic(err)
		}
	} else {
		stat, err := index.GetStats()
		if err != nil {
			panic(err)
		}
		if stat.NumberOfDocuments == 0 {
			shouldCreateIndex = true
		}
	}

	client := &Client{
		search:       meili,
		q:            query,
		subject:      "subjects",
		subjectIndex: meili.Index("subjects"),
		log:          log.Named("search"),
		subjectRepo:  subjectRepo,
	}

	if shouldCreateIndex {
	}
	go firstRun(client)

	return client, nil
}

type Client struct {
	subjectRepo  domain.SubjectRepo
	search       *meilisearch.Client
	q            *query.Query
	subjectIndex *meilisearch.Index
	log          *zap.Logger
	subject      string
}

// OnSubjectUpdate is the hook called by canal.
func (c *Client) OnSubjectUpdate(ctx context.Context, id model.SubjectID) error {
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
func (c *Client) OnSubjectDelete(ctx context.Context, id model.SubjectID) error {
	_, err := c.subjectIndex.DeleteDocument(strconv.FormatUint(uint64(id), 10))

	return errgo.Wrap(err, "search")
}

// UpsertSubject add subject to search backend.
func (c *Client) upsertSubject(ctx context.Context, s subjectIndex) error {
	_, err := c.subjectIndex.UpdateDocuments(s)

	return errgo.Wrap(err, "search")
}

func (c *Client) DeleteSubject(ctx context.Context, id model.SubjectID) error {
	_, err := c.subjectIndex.Delete(strconv.FormatUint(uint64(id), 10))

	return errgo.Wrap(err, "delete")
}

func firstRun(client *Client) {
	client.log.Info("search initialize")
	_, err := client.search.CreateIndex(&meilisearch.IndexConfig{
		Uid:        "subjects",
		PrimaryKey: "id",
	})
	if err != nil {
		panic(err)
	}
	subjectIndex := client.search.Index("subjects")

	_, err = subjectIndex.UpdateFilterableAttributes(&[]string{
		"type",
		"score",
		"nsfw",
		"rank",
		"date",
		"tag",
	})
	if err != nil {
		panic(err)
	}

	_, err = subjectIndex.UpdateSearchableAttributes(&[]string{
		"name",
		"name_cn",
		"summary",
	})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// subjects, err := client.q.Subject.WithContext(ctx).Find()
	// if err != nil {
	// 	panic(err)
	// }
	//
	// for _, subject := range subjects {
	// 	err := client.OnSubjectUpdate(ctx, subject.ID)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	//
	// }
	//
	// return

	maxSubject, err := client.q.Subject.WithContext(ctx).Limit(1).Order(client.q.Subject.ID.Desc()).First()
	if err != nil {
		panic(err)
	}

	client.log.Info(fmt.Sprintln("run full search index with max subject id", maxSubject))

	for i := model.SubjectID(1); i < maxSubject.ID; i++ {
		err := client.OnSubjectUpdate(ctx, i)
		if err != nil {
			client.log.Error("error when updating subject", zap.Error(err))
		}
	}
}
