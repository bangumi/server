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
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/errgo"
)

func GetQuery(tb testing.TB) *query.Query {
	tb.Helper()
	db, err := newGorm(tb, config.NewAppConfig())
	require.NoError(tb, err)

	return query.Use(db)
}

func GetGorm(tb testing.TB) *gorm.DB {
	tb.Helper()
	db, err := newGorm(tb, config.NewAppConfig())
	require.NoError(tb, err)

	return db
}

func newGorm(tb testing.TB, c config.AppConfig) (*gorm.DB, error) {
	tb.Helper()
	conn, err := driver.NewMysqlConnectionPool(c)
	if err != nil {
		return nil, errgo.Wrap(err, "sql.Open")
	}

	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conn, DisableDatetimePrecision: true}), &gorm.Config{
		Logger: gormLogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormLogger.Config{
				LogLevel:                  gormLogger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
		QueryFields: true,
		PrepareStmt: true,
	})
	require.NoError(tb, err)

	return db, errgo.Wrap(err, "gorm.Open")
}
