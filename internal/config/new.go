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
	"os"
	"strings"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/gookit/goutil/dump"
	"github.com/ilyakaznacheev/cleanenv"

	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
)

//nolint:gochecknoglobals
var config = pflag.String("config", "", "config file location, optional")

//nolint:gochecknoinits
func init() {
	pflag.Parse()
}

type AppConfig struct {
	Debug             map[string]bool
	RedisURL          string `yaml:"redis_url" env:"REDIS_URI" env-default:"redis://127.0.0.1:6379/0"`
	MySQLHost         string `yaml:"mysql_host" env:"MYSQL_HOST" env-default:"127.0.0.1"`
	MySQLPort         string `yaml:"mysql_port" env:"MYSQL_PORT" env-default:"3306"`
	MySQLUserName     string `yaml:"mysql_user" env:"MYSQL_USER" env-default:"user"`
	MySQLPassword     string `yaml:"mysql_pass" env:"MYSQL_PASS" env-default:"password"`
	MySQLDatabase     string `yaml:"mysql_db" env:"MYSQL_DB" env-default:"bangumi"`
	HCaptchaSecretKey string `yaml:"hcaptcha_secret_key" env:"HCAPTCHA_SECRET_KEY" env-default:"0x0000000000000000000000000000000000000000"`
	FrontendDomain    string `yaml:"web_domain" env:"WEB_DOMAIN"` // new frontend web page domain, like next.bgm.tv
	HTTPHost          string `yaml:"http_host" env:"HTTP_HOST" env-default:"127.0.0.1"`
	HTTPPort          int    `env:"HTTP_PORT" env-default:"3000"`
	MySQLMaxConn      int    `yaml:"mysql_max_conn" env:"MYSQL_MAX_CONNECTION" env-default:"4"`

	NsfwWord     string `yaml:"nsfw_word"`
	DisableWords string `yaml:"disable_words"`
	BannedDomain string `yaml:"banned_domain"`
}

func NewAppConfig() (AppConfig, error) {
	logger.Info("reading app config", zap.String("version", Version))
	var debug = make(map[string]bool)
	if debugV := os.Getenv("DEBUG"); debugV != "" {
		logger.Info("enable debug: " + debugV)
		for _, v := range strings.Split(debugV, ",") {
			v = strings.TrimSpace(v)
			debug[v] = true
		}
	}

	var cfg AppConfig
	var err error
	if *config != "" {
		logger.Info("reading app config file", zap.Stringp("config", config))
		err = errgo.Wrap(cleanenv.ReadConfig(*config, &cfg), "ReadConfig")
	} else {
		err = errgo.Wrap(cleanenv.ReadEnv(&cfg), "ReadEnv")
	}

	if err != nil {
		return AppConfig{}, err
	}

	dump.P(cfg)

	cfg.Debug = debug

	return cfg, nil

}
