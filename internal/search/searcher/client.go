package searcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/avast/retry-go/v5"
	wiki "github.com/bangumi/wiki-parser-go"
	"github.com/labstack/echo/v4"
	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
)

type Searcher interface {
	Handle(c echo.Context) error

	OnAdded(ctx context.Context, id uint32) error
	OnUpdate(ctx context.Context, id uint32) error
	OnDelete(ctx context.Context, id uint32) error
}

type Document interface {
	GetID() string
}

func NeedFirstRun(meili meilisearch.ServiceManager, idx string) (bool, error) {
	if os.Getenv("CHII_SEARCH_INIT") == "true" {
		return true, nil
	}

	index, err := meili.GetIndex(idx)
	if err != nil {
		var e *meilisearch.Error
		if errors.As(err, &e) {
			return true, nil
		}
		return false, errgo.Wrap(err, fmt.Sprintf("get index %s", idx))
	}

	stat, err := index.GetStats()
	if err != nil {
		return false, errgo.Wrap(err, fmt.Sprintf("get index %s stats", idx))
	}

	return stat.NumberOfDocuments == 0, nil
}

func ValidateConfigs(cfg config.AppConfig) error {
	return nil
}

func ExtractAliases(w wiki.Wiki) []string {
	aliases := []string{}
	for _, field := range w.Fields {
		if field.Key == "中文名" {
			aliases = append(aliases, GetWikiValues(field)...)
		}
		if field.Key == "简体中文名" {
			aliases = append(aliases, GetWikiValues(field)...)
		}
	}
	for _, field := range w.Fields {
		if field.Key == "别名" {
			aliases = append(aliases, GetWikiValues(field)...)
		}
	}
	return aliases
}

func GetWikiValues(f wiki.Field) []string {
	if f.Null {
		return nil
	}

	if !f.Array {
		return []string{f.Value}
	}

	var s = make([]string, len(f.Values))
	for i, value := range f.Values {
		s[i] = value.Value
	}
	return s
}

func NewSendBatch(log *zap.Logger, index meilisearch.IndexManager) func([]Document) {
	var retrier = retry.New(
		retry.OnRetry(func(n uint, err error) {
			log.Warn("failed to send batch", zap.Uint("attempt", n), zap.Error(err))
		}),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(time.Second),
		retry.Attempts(5), //nolint:mnd
		retry.RetryIf(func(err error) bool {
			var r = &meilisearch.Error{}
			return errors.As(err, &r)
		}),
	)

	return func(items []Document) {
		log.Debug("send batch to meilisearch", zap.Int("len", len(items)))
		err := retrier.Do(func() error {
			_, err := index.UpdateDocuments(items, &meilisearch.DocumentOptions{PrimaryKey: lo.ToPtr("id")})
			return err
		})
		if err != nil {
			log.Error("failed to send batch", zap.Error(err))
		}
	}
}

func NewDedupeFunc() func([]Document) []Document {
	return func(items []Document) []Document {
		mutable.Reverse(items)
		return lo.UniqBy(items, func(item Document) string {
			return item.GetID()
		})
	}
}

func GetAttributes(rt reflect.Type, tag string) *[]string {
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

func InitIndex(log *zap.Logger, meili meilisearch.ServiceManager, idx string, rt reflect.Type, rankRule *[]string) {
	_, err := meili.CreateIndex(&meilisearch.IndexConfig{
		Uid:        idx,
		PrimaryKey: "id",
	})
	if err != nil {
		log.Fatal("failed to create search index", zap.Error(err))
		return
	}

	index := meili.Index(idx)

	log.Info("set sortable attributes", zap.Strings("attributes", *GetAttributes(rt, "sortable")))
	_, err = index.UpdateSortableAttributes(GetAttributes(rt, "sortable"))
	if err != nil {
		log.Fatal("failed to update search index sortable attributes", zap.Error(err))
		return
	}

	log.Info("set filterable attributes", zap.Strings("attributes", *GetAttributes(rt, "filterable")))
	_, err = index.UpdateFilterableAttributes(lo.ToPtr(
		lo.Map(*GetAttributes(rt, "filterable"), func(s string, index int) any {
			return s
		})))
	if err != nil {
		log.Fatal("failed to update search index filterable attributes", zap.Error(err))
		return
	}

	log.Info("set searchable attributes", zap.Strings("attributes", *GetAttributes(rt, "searchable")))
	_, err = index.UpdateSearchableAttributes(GetAttributes(rt, "searchable"))
	if err != nil {
		log.Fatal("failed to update search index searchable attributes", zap.Error(err))
		return
	}

	log.Info("set ranking rules", zap.Strings("rule", *rankRule))
	_, err = index.UpdateRankingRules(rankRule)
	if err != nil {
		log.Fatal("failed to update search index searchable attributes", zap.Error(err))
		return
	}
}
