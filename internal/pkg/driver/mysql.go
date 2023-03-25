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
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/pkg/logger"
)

var setLoggerOnce = sync.Once{}

func NewMysqlConnectionPool(c config.AppConfig) (*sql.DB, error) {
	setLoggerOnce.Do(func() {
		_ = mysql.SetLogger(logger.StdAt(zap.ErrorLevel))
	})

	logger.Infoln("creating sql connection pool with size", c.Mysql.MaxConn)

	u := mysql.NewConfig()
	u.User = c.Mysql.UserName
	u.Passwd = c.Mysql.Password
	u.Net = "tcp"
	u.Addr = net.JoinHostPort(c.Mysql.Host, c.Mysql.Port)
	u.DBName = c.Mysql.Database
	u.Loc = time.UTC
	u.ParseTime = true

	db, err := sql.Open("mysql", u.FormatDSN())
	if err != nil {
		return nil, errgo.Wrap(err, "mysql: failed to create sql connection pool")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, errgo.Wrap(err, "mysql: failed to ping")
	}

	db.SetMaxOpenConns(c.Mysql.MaxConn)
	db.SetConnMaxIdleTime(c.Mysql.MaxIdleTime)
	db.SetConnMaxLifetime(c.Mysql.MaxLifeTime)

	return db, nil
}
