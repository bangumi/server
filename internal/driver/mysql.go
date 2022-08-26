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

package driver

import (
	"context"
	"database/sql"
	"net"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
)

var setLoggerOnce = sync.Once{}

func NewMysqlConnectionPool(c config.AppConfig) (*sql.DB, error) {
	const maxIdleTime = time.Hour * 6

	setLoggerOnce.Do(func() {
		_ = mysql.SetLogger(logger.StdAt(zap.ErrorLevel))
	})

	logger.Infoln("creating sql connection pool with size", c.MySQLMaxConn)

	u := mysql.NewConfig()
	u.User = c.MySQLUserName
	u.Passwd = c.MySQLPassword
	u.Net = "tcp"
	u.Addr = net.JoinHostPort(c.MySQLHost, c.MySQLPort)
	u.DBName = c.MySQLDatabase
	u.Loc = time.UTC
	u.ParseTime = true

	db, err := sql.Open("mysql", u.FormatDSN())
	if err != nil {
		return nil, errgo.Wrap(err, "failed to create sql connection pool")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, errgo.Wrap(err, "db.PingContext")
	}

	db.SetMaxOpenConns(c.MySQLMaxConn)
	// default mysql has 7 hour timeout
	db.SetConnMaxIdleTime(maxIdleTime)

	return db, nil
}
