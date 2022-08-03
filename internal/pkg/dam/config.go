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

package dam

import (
	"regexp"

	"github.com/bangumi/server/internal/pkg/errgo"
)

type Config struct {
	NsfwWord     string `json:"nsfw_word"`
	DisableWords string `json:"disable_words"`
	BannedDomain string `json:"banned_domain"`
}

func (c *Config) Load() (*Dam, error) {
	var cc Dam
	var err error
	cc.nsfwWord, err = regexp.CompilePOSIX(c.NsfwWord)
	if err != nil {
		return nil, errgo.Wrap(err, "nsfw_word")
	}

	cc.disableWord, err = regexp.CompilePOSIX(c.DisableWords)
	if err != nil {
		return nil, errgo.Wrap(err, "disable_words")
	}

	cc.bannedDomain, err = regexp.CompilePOSIX(c.BannedDomain)
	if err != nil {
		return nil, errgo.Wrap(err, "banned_domain")
	}

	return &cc, nil
}
