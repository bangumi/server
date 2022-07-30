package search

import (
	"context"
	"net/url"
	"strconv"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/config"
)

func New(c config.AppConfig, db *gorm.DB) (*Client, error) {
	u, err := url.Parse("")
	if err != nil {
		return nil, errors.Wrap(err, "url")
	}

	var options = []elastic.ClientOptionFunc{
		elastic.SetSniff(false),
		elastic.SetGzip(false),
	}
	if u.User != nil {
		p, _ := u.User.Password()
		options = append(options, elastic.SetBasicAuth(u.User.Username(), p))
	}

	client, err := elastic.NewClient(options...)
	if err != nil {
		return nil, errors.Wrap(err, "init es client")
	}

	return &Client{es: client, db: db, subject: "subjects"}, nil
}

type Client struct {
	es      *elastic.Client
	db      *gorm.DB
	subject string
}

// UpsertSubject add subject to search backend.
func (c *Client) UpsertSubject(ctx context.Context, s Subject) error {
	_, err := c.es.Update().Index(c.subject).Id(strconv.Itoa(int(s.Record.ID))).
		DocAsUpsert(true).Doc(s).Do(ctx)

	return errors.Wrap(err, "es")
}

func (c *Client) DeleteSubject(ctx context.Context, id string) error {
	_, err := c.es.Delete().Index(c.subject).Id(id).Do(ctx)

	return errors.Wrap(err, "delete")
}
