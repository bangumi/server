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

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/logger"
)

const defaultMaxMysqlConnection = 4

func NewAppConfig() AppConfig {
	logger.Info("reading app config", zap.String("version", Version))
	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "user")
	pass := getEnv("MYSQL_PASS", "password")
	db := getEnv("MYSQL_DB", "bangumi")
	maxConnection := getEnvInt("MYSQL_MAX_CONNECTION", defaultMaxMysqlConnection)

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

	redisURL := getEnv("REDIS_URI", "redis://127.0.0.1:6379/0")
	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Fatal("failed to parse redis url", zap.String("url", redisURL))
	}

	return AppConfig{
		RedisOptions:  redisOptions,
		MySQLHost:     host,
		MySQLPort:     port,
		MySQLUserName: user,
		MySQLPassword: pass,
		MySQLDatabase: db,
		MySQLMaxConn:  maxConnection,
		Debug:         debug,
		HTTPPort:      httpPort,

		HCaptchaSecretKey: getEnv("HCAPTCHA_SECRET_KEY", ""),
		FrontendDomain:    getEnv("WEB_DOMAIN", ""),
	}
}

type AppConfig struct {
	Debug             map[string]bool
	RedisOptions      *redis.Options
	MySQLHost         string
	MySQLPort         string
	MySQLUserName     string
	MySQLPassword     string
	MySQLDatabase     string
	HCaptchaSecretKey string
	FrontendDomain    string // new frontend web page domain, like next.bgm.tv
	MySQLMaxConn      int
	HTTPPort          int
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
