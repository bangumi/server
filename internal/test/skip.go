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
	"os"
	"strconv"
	"strings"
)

const EnvMysql = "mysql"
const EnvRedis = "redis"

// RequireEnv
//  func TestGet(t *testing.T) {
//    RequireEnv(t, test.EnvRedis, test.EnvMysql)
//    ...
//  }
func RequireEnv(t TB, envs ...string) {
	t.Helper()
	for _, env := range envs {
		if v, ok := os.LookupEnv(strings.ToUpper("test_" + env)); !ok {
			t.SkipNow()
		} else if ok, _ = strconv.ParseBool(v); !ok {
			t.SkipNow()
		}
	}
}
