package search

import (
	"context"
	"net/url"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func New(c config.AppConfig, db *gorm.DB) (*Client, error) {
	if _, err := url.Parse(c.MeiliSearchURL); err != nil {
		return nil, errgo.Wrap(err, "url.Parse")
	}

	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    c.MeiliSearchURL,
		APIKey:  "masterKey",
		Timeout: time.Second * 5,
	})

	return &Client{search: client, db: db, subject: "subjects"}, nil
}

type Client struct {
	search  *meilisearch.Client
	db      *gorm.DB
	subject string
}

// UpsertSubject add subject to search backend.
func (c *Client) UpsertSubject(ctx context.Context, s Subject) error {
	_, err := c.search.Index(c.subject).UpdateDocuments(s, "id")

	return errgo.Wrap(err, "search")
}

func (c *Client) DeleteSubject(ctx context.Context, id string) error {
	_, err := c.search.Index(c.subject).Delete(id)

	return errgo.Wrap(err, "delete")
}
