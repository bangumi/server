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

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/pkg/logger"
)

const defaultMaxMysqlConnection = 4

func NewAppConfig() AppConfig {
	logger.Info("reading app config", zap.String("version", Version))
	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "user")
	pass := getEnv("MYSQL_PASS", "password")
	db := getEnv("MYSQL_DB", "bangumi")

	httpPort, err := strconv.Atoi(getEnv("HTTP_PORT", "3000"))
	if err != nil {
		logger.Fatal("can't parse http port", zap.Error(err))
	}

	var debug = make(map[string]bool)
	if debugV := getEnv("DEBUG", ""); debugV != "" {
		logger.Info("enable debug: " + debugV)
		for _, v := range strings.Split(debugV, ",") {
			v = strings.TrimSpace(v)
			debug[v] = true
		}
	}

	return AppConfig{
		RedisURL:      getEnv("REDIS_URI", "redis://127.0.0.1:6379/0"),
		MySQLHost:     host,
		MySQLPort:     port,
		MySQLUserName: user,
		MySQLPassword: pass,
		MySQLDatabase: db,
		MySQLMaxConn:  getEnvInt("MYSQL_MAX_CONNECTION", defaultMaxMysqlConnection),
		Debug:         debug,
		HTTPPort:      httpPort,

		HCaptchaSecretKey: getEnv("HCAPTCHA_SECRET_KEY", "0x0000000000000000000000000000000000000000"),
		WebDomain:         getEnv("WEB_DOMAIN", ""),
		HTTPHost:          getEnv("HTTP_HOST", "127.0.0.1"),
	}
}

type AppConfig struct {
	Debug             map[string]bool
	RedisURL          string
	MySQLHost         string
	MySQLPort         string
	MySQLUserName     string
	MySQLPassword     string
	MySQLDatabase     string
	HCaptchaSecretKey string
	WebDomain         string // new frontend web page domain, like next.bgm.tv
	HTTPHost          string
	HTTPPort          int
	MySQLMaxConn      int
}

func getEnv(n, v string) string {
	if e, ok := os.LookupEnv(n); ok {
		return e
	}

	return v
}

func getEnvInt(name string, defaultValue int) int {
	if raw, ok := os.LookupEnv(name); ok {
		v, err := strconv.Atoi(raw)
		if err != nil {
			logger.Fatal(fmt.Sprintf("failed to read config from env, can't convert '%v' to int", raw),
				zap.Error(err), zap.String("env_name", name))
		}

		return v
	}

	return defaultValue
}
