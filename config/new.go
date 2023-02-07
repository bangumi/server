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
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/pkg/logger"
)

func AppConfigReader(appType string) func() (AppConfig, error) {
	return func() (AppConfig, error) {
		c, err := NewAppConfig()
		c.AppType = appType
		return c, err
	}
}

func NewAppConfig() (AppConfig, error) {
	cli := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	var config = cli.String("config", "", "config file location, optional")
	_ = cli.Parse(os.Args[1:])

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

	return cfg, nil
}
