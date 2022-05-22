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

package dal

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
)

const maxIdleTime = time.Hour * 6
const slowQueryTimeout = time.Millisecond * 200

func NewDB(
	conn *sql.DB, c config.AppConfig, scope tally.Scope, register prometheus.Registerer,
) (*gorm.DB, error) {
	var gLog gormLogger.Interface
	if c.Debug["gorm"] {
		logger.Info("enable gorm debug mode, will log all sql")
		gLog = gormLogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormLogger.Config{
				LogLevel:                  gormLogger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)
	} else {
		gLog = gormLogger.New(
			logger.Std(),
			gormLogger.Config{
				SlowThreshold:             slowQueryTimeout,
				LogLevel:                  gormLogger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conn,
		DisableDatetimePrecision: true,
	}), &gorm.Config{
		Logger:      gLog,
		QueryFields: true,
		PrepareStmt: true,
	})
	if err != nil {
		return nil, errgo.Wrap(err, "create dal")
	}

	if err = setupMetrics(db, scope, register); err != nil {
		return nil, errgo.Wrap(err, "setup metrics")
	}

	if c.Debug["gorm"] {
		return db.Debug(), errgo.Wrap(err, "init gorm")
	}

	return db, errgo.Wrap(err, "init gorm")
}

func NewConnectionPool(c config.AppConfig) (*sql.DB, error) {
	logger.Infoln("creating sql connection pool with size", c.MySQLMaxConn)
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=UTC",
			c.MySQLUserName, c.MySQLPassword, c.MySQLHost, c.MySQLPort, c.MySQLDatabase))
	if err != nil {
		return nil, errgo.Wrap(err, "failed to create sql connection pool")
	}
	db.SetMaxOpenConns(c.MySQLMaxConn)
	// default mysql has 7 hour timeout
	db.SetConnMaxIdleTime(maxIdleTime)

	return db, nil
}
