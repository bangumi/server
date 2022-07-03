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
	"testing"
)

const EnvMysql = "mysql"
const EnvRedis = "redis"
const EnvExternalHTTP = "http" // external http server like hCaptcha

const TreeHoleAccessToken = "a_development_access_token"

// RequireEnv
//  func TestGet(t *testing.T) {
//    RequireEnv(t, test.EnvRedis, test.EnvMysql)
//    ...
//  }
func RequireEnv(tb testing.TB, envs ...string) {
	tb.Helper()
	for _, env := range envs {
		if v, ok := os.LookupEnv(strings.ToUpper("test_" + env)); !ok {
			tb.SkipNow()
		} else if ok, _ = strconv.ParseBool(v); !ok {
			tb.SkipNow()
		}
	}
}
