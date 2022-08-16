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

func New(c config.AppConfig, subjectRepo domain.SubjectRepo, log *zap.Logger, query *query.Query) (*Client, error) {
	if _, err := url.Parse(c.MeiliSearchURL); err != nil {
		return nil, errgo.Wrap(err, "url.Parse")
	}

	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    c.MeiliSearchURL,
		APIKey:  c.MeiliSearchKey,
		Timeout: time.Second,
	})

	_, err := client.GetVersion()
	if err != nil {
		return nil, errgo.Wrap(err, "meilisearch")
	}

	return &Client{
		search:      client,
		q:           query,
		subject:     "subjects",
		log:         log.Named("search"),
		subjectRepo: subjectRepo,
	}, nil
}

type Client struct {
	subjectRepo domain.SubjectRepo
	search      *meilisearch.Client
	q           *query.Query
	log         *zap.Logger
	subject     string
}

// OnSubjectUpdate is the hook called by canal.
func (c *Client) OnSubjectUpdate(ctx context.Context, id model.SubjectID) error {
	s, err := c.subjectRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.DeleteSubject(ctx, strconv.Itoa(int(id)))
		}

		c.log.Error("unexpected error get subject from mysql", zap.Error(err), log.SubjectID(id))
		return errgo.Wrap(err, "subjectRepo.Get")
	}

	extracted := extractSubject(&s)

	return c.upsertSubject(ctx, extracted)
}

// UpsertSubject add subject to search backend.
func (c *Client) upsertSubject(ctx context.Context, s subjectIndex) error {
	_, err := c.search.Index(c.subject).UpdateDocuments(s, "id")

	return errgo.Wrap(err, "search")
}

func (c *Client) DeleteSubject(ctx context.Context, id string) error {
	_, err := c.search.Index(c.subject).Delete(id)

	return errgo.Wrap(err, "delete")
}
