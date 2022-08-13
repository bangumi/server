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

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
)

//nolint:govet
type AppConfig struct {
	Debug map[string]bool `yaml:"debug"`

	RedisURL            string `yaml:"redis_url" env:"REDIS_URI" env-default:"redis://127.0.0.1:6379/0"`
	MySQLHost           string `yaml:"mysql_host" env:"MYSQL_HOST" env-default:"127.0.0.1"`
	MySQLPort           string `yaml:"mysql_port" env:"MYSQL_PORT" env-default:"3306"`
	MySQLUserName       string `yaml:"mysql_user" env:"MYSQL_USER" env-default:"user"`
	MySQLBinlogServerID int    `yaml:"mysql_slave_id" env:"MYSQL_SLAVE_ID" env-default:"4"`
	MySQLPassword       string `yaml:"mysql_pass" env:"MYSQL_PASS" env-default:"password"`
	MySQLDatabase       string `yaml:"mysql_db" env:"MYSQL_DB" env-default:"bangumi"`
	MySQLMaxConn        int    `yaml:"mysql_max_connection" env:"MYSQL_MAX_CONNECTION" env-default:"4"`

	WebDomain string `yaml:"web_domain" env:"WEB_DOMAIN"` // new frontend web page domain
	HTTPHost  string `yaml:"http_host" env:"HTTP_HOST" env-default:"127.0.0.1"`
	HTTPPort  int    `yaml:"http_port" env:"HTTP_PORT" env-default:"3000"`

	KafkaBroker string `yaml:"kafka_broker" env:"KAFKA_BROKER"`

	HCaptchaSecretKey string `yaml:"hcaptcha_secret_key" env:"HCAPTCHA_SECRET_KEY"`

	NsfwWord     string `yaml:"nsfw_word"`
	DisableWords string `yaml:"disable_words"`
	BannedDomain string `yaml:"banned_domain"`
}

func NewAppConfig() (AppConfig, error) {
	cli := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	var config = cli.String("config", "", "config file location, optional")
	_ = cli.Parse(os.Args[1:])

	logger.Info("reading app config", zap.String("version", Version))
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

	// 太长了
	cfg.HCaptchaSecretKey = setDefault(cfg.HCaptchaSecretKey, "0x0000000000000000000000000000000000000000")

	return cfg, nil
}

func setDefault(s string, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}
