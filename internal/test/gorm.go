// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/errgo"
)

func GetGorm(t TB) *gorm.DB {
	t.Helper()
	db, err := newGorm(t, config.NewAppConfig())
	require.NoError(t, err)

	return db
}

func newGorm(t TB, c config.AppConfig) (*gorm.DB, error) {
	conn, err := dal.NewConnectionPool(c)
	if err != nil {
		return nil, errgo.Wrap(err, "sql.Open")
	}

	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conn, DisableDatetimePrecision: true}))
	require.NoError(t, err)
	db, err = dal.NewDB(conn, c, tally.NoopScope, prometheus.NewRegistry())

	return db, errgo.Wrap(err, "gorm.Open")
}
