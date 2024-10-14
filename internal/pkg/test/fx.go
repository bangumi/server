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
	"encoding/json"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/dal"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/driver"
)

func Fx(t testing.TB, target ...fx.Option) {
	t.Helper()
	err := fx.New(
		append(target, fx.NopLogger,

			// driver and connector
			fx.Provide(
				config.AppConfigReader(config.AppTypeHTTP),
				driver.NewRedisClient,   // redis
				driver.NewRueidisClient, // redis
				driver.NewMysqlSqlDB,    // mysql
				func() *resty.Client {
					httpClient := resty.New().SetJSONEscapeHTML(false)
					httpClient.JSONUnmarshal = json.Unmarshal
					httpClient.JSONMarshal = json.Marshal
					return httpClient
				},
			),

			dal.Module,

			fx.Provide(cache.NewRedisCache, zap.NewNop),
		)...,
	).Err()

	require.NoError(t, err)
}
