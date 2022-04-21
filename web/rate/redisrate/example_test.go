//nolint:goheader
package redisrate_test

import (
	"context"
	"fmt"

	"github.com/go-redis/redis_rate/v9"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/driver"
)

func ExampleNewLimiter() {
	ctx := context.Background()

	rdb, err := driver.NewRedisClient(config.NewAppConfig())
	if err != nil {
		panic(err)
	}

	_ = rdb.FlushDB(ctx).Err()

	limiter := redis_rate.NewLimiter(rdb)
	res, err := limiter.Allow(ctx, "project:123", redis_rate.PerSecond(10))
	if err != nil {
		panic(err)
	}
	fmt.Println("allowed", res.Allowed, "remaining", res.Remaining)
	// Output: allowed 1 remaining 9
}
