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

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/gookit/goutil/dump"
	"github.com/ilyakaznacheev/cleanenv"

	"github.com/bangumi/server/internal/pkg/logger"
)

const defaultMaxMysqlConnection = 4

//nolint:gochecknoglobals
var config = pflag.String("config", "", "config file location, optional")

//nolint:gochecknoinits
func init() {
	pflag.Parse()
}

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

	var cfg AppConfig
	if *config == "" {
		err := cleanenv.ReadConfig(*config, &cfg)
		if err != nil {
			panic(err)
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(err)
	}

	dump.P(cfg)

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
		FrontendDomain:    getEnv("WEB_DOMAIN", ""),
		HTTPHost:          getEnv("HTTP_HOST", "127.0.0.1"),
	}
}

type AppConfig struct {
	Debug             map[string]bool
	RedisURL          string `json:"redis_url" env:"REDIS_URI" env-default:"redis://127.0.0.1:6379/0"`
	MySQLHost         string `json:"mysql_host" env:"MYSQL_HOST" env-default:"127.0.0.1"`
	MySQLPort         string `json:"mysql_port" env:"MYSQL_PORT" env-default:"3306"`
	MySQLUserName     string `json:"mysql_user" env:"MYSQL_USER" env-default:"user"`
	MySQLPassword     string `json:"mysql_pass" env:"MYSQL_PASS" env-default:"password"`
	MySQLDatabase     string `json:"mysql_db" env:"MYSQL_DB" env-default:"bangumi"`
	HCaptchaSecretKey string
	FrontendDomain    string // new frontend web page domain, like next.bgm.tv
	HTTPHost          string `json:"http_host" env:"HTTP_HOST" env-default:"127.0.0.1"`
	HTTPPort          int    `env:"HTTP_PORT" env-default:"3000"`
	MySQLMaxConn      int    `json:"mysql_max_conn" env:"MYSQL_MAX_CONNECTION" env-default:"4"`
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
