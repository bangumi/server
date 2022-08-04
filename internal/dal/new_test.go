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

package dal_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally/v4"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestNewDB(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvMysql)
	cfg, err := config.NewAppConfig()
	require.NoError(t, err)

	conn, err := driver.NewMysqlConnectionPool(cfg)
	require.NoError(t, err)
	db, err := dal.NewDB(conn, cfg, tally.NoopScope, prometheus.NewRegistry())
	require.NoError(t, err)

	err = db.Exec("select 0;").Error
	require.NoError(t, err)
}
