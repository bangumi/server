package cron

import (
	"context"
	"strconv"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
)

type C struct {
	c *cron.Cron
}

func Start() error {
	var c *C
	err := fx.New(
		dal.Module,
		fx.Provide(config.NewAppConfig),
		fx.Provide(search.New),
		fx.Provide(New),
		fx.Populate(&c),
		fx.NopLogger,
	).Err()
	if err != nil {
		return errors.Wrap(err, "dependency inject")
	}
	logger.Info("start cron daemon")
	c.Run()

	return nil
}

func New(db *gorm.DB, es *search.Client) (*C, error) {
	logger.Info("creating cron")
	c := cron.New(cron.WithLogger(cron.DefaultLogger))

	q := query.Use(db)

	_, err := c.AddFunc("@daily", indexMissingSubject(q, es))
	if err != nil {
		return nil, errors.Wrap(err, "add missing subject finder")
	}

	return &C{c}, nil
}

func (c *C) Run() {
	c.c.Run()
}

func deadline1s() (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), time.Now().Add(time.Second))
}

func indexMissingSubject(q *query.Query, es *search.Client) func() {
	return func() {
		s, err := q.Subject.WithContext(context.Background()).Limit(1).Order(q.Subject.ID.Desc()).First()
		if err != nil {
			logger.Error("can't get maximum subject ID", zap.Error(err))

			return
		}

		logger.Info("max subject id " + strconv.Itoa(int(s.ID)))

		subjectRepo, err := subject.NewMysqlRepo(q, logger.Copy())
		if err != nil {
			panic(err)
		}

		for i := model.SubjectID(1); i < s.ID; i++ {
			time.Sleep(time.Millisecond * 200)
			result, err := subjectRepo.Get(context.Background(), i)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					continue
				}
				logger.Error("failed to get subject from repo", zap.Error(err), log.SubjectID(i))
			}

			if err = onSubject(es, &result); err != nil {
				logger.Error("error on index subject", zap.Error(err), log.SubjectID(i))
			}
		}
		logger.Info("finish full indexing")
	}
}

func onSubject(es *search.Client, s *model.Subject) error {
	if s.Redirect != 0 {
		ctx, cancel := deadline1s()
		defer cancel()
		if err := es.DeleteSubject(ctx, strconv.Itoa(int(s.ID))); err != nil {
			if !elastic.IsNotFound(err) {
				return errors.Wrap(err, "delete")
			}
		}

		return nil
	}

	ctx, cancel := deadline1s()
	defer cancel()
	doc, err := es.ExtractSubject(ctx, s)
	if err != nil {
		return errors.Wrap(err, "upsert")
	}

	ctx, cancel = deadline1s()
	defer cancel()
	if err := es.UpsertSubject(ctx, doc); err != nil {
		return errors.Wrap(err, "upsert")
	}

	return nil
}
