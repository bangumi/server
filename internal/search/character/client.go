package character

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/meilisearch/meilisearch-go"
	"github.com/trim21/errgo"
	"github.com/trim21/pkg/queue"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/search/searcher"
)

const (
	idx = "characters"
)

func New(
	cfg config.AppConfig,
	meili meilisearch.ServiceManager,
	repo character.Repo,
	log *zap.Logger,
	query *query.Query,
) (searcher.Searcher, error) {
	if repo == nil {
		return nil, fmt.Errorf("nil characterRepo")
	}
	c := &client{
		meili: meili,
		repo:  repo,
		index: meili.Index(idx),
		log:   log.Named("search").With(zap.String("index", idx)),
		q:     query,
	}

	if cfg.AppType != config.AppTypeCanal {
		return c, nil
	}

	return c, c.canalInit(cfg)
}

type client struct {
	repo  character.Repo
	index meilisearch.IndexManager

	meili meilisearch.ServiceManager
	log   *zap.Logger
	q     *query.Query

	queue *queue.Batched[searcher.Document]
}

func (c *client) Close() {
	if c.queue != nil {
		c.queue.Close()
	}
}

func (c *client) canalInit(cfg config.AppConfig) error {
	if err := searcher.ValidateConfigs(cfg); err != nil {
		return errgo.Wrap(err, "validate search config")
	}
	c.queue = searcher.NewBatchQueue(cfg, c.log, c.index)
	searcher.RegisterQueueMetrics(idx, c.queue)
	shouldCreateIndex, err := searcher.NeedFirstRun(c.meili, idx)
	if err != nil {
		return err
	}
	if shouldCreateIndex {
		go c.firstRun()
	}
	return nil
}

//nolint:funlen
func (c *client) firstRun() {
	c.log.Info("search initialize")
	rt := reflect.TypeOf(document{})
	searcher.InitIndex(c.log, c.meili, idx, rt, rankRule())

	ctx := context.Background()

	maxItem, err := c.q.Character.WithContext(ctx).Limit(1).Order(c.q.Character.ID.Desc()).Take()
	if err != nil {
		c.log.Fatal("failed to get current max id", zap.Error(err))
		return
	}

	c.log.Info(fmt.Sprintf("run full search index with max %s id %d", idx, maxItem.ID))

	width := len(strconv.Itoa(int(maxItem.ID)))
	for i := model.CharacterID(1); i <= maxItem.ID; i++ {
		if i%10000 == 0 {
			c.log.Info(fmt.Sprintf("progress %*d/%d", width, i, maxItem.ID))
		}

		err := c.OnUpdate(ctx, i)
		if err != nil {
			c.log.Error("error when updating", zap.Error(err))
		}
	}
}
