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

package test

import (
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/pkg/driver"
)

func GetRedis(tb testing.TB) *redis.Client {
	tb.Helper()

	cfg, err := config.NewAppConfig()
	require.NoError(tb, err)
	db, err := driver.NewRedisClient(cfg)
	require.NoError(tb, err)

	return db
}
