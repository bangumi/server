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

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/bangumi/server/internal/pkg/errgo"
)

type File struct {
	NsfwWord     string `yaml:"nsfw_word"`
	DisableWords string `yaml:"disable_words"`
	BannedDomain string `yaml:"banned_domain"`
}

func ReadFileConfig() (File, error) {
	config := pflag.String("config", "", "")
	pflag.Parse()
	var cfg File
	if *config != "" {
		f, err := os.Open(*config)
		if err != nil {
			return File{}, errgo.Wrap(err, "os.Open")
		}
		defer f.Close()

		err = yaml.NewDecoder(f).Decode(&cfg)
		if err != nil {
			return File{}, errgo.Wrap(err, "toml.Decode")
		}
	}

	return cfg, nil
}
