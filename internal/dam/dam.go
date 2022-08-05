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
	"strings"
	"unicode"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/pkg/errgo"
)

type Dam struct {
	nsfwWord     *regexp.Regexp
	disableWord  *regexp.Regexp
	bannedDomain *regexp.Regexp
}

func New(c config.AppConfig) (Dam, error) {
	var cc Dam
	var err error
	if c.NsfwWord != "" {
		cc.nsfwWord, err = regexp.Compile(c.NsfwWord)
		if err != nil {
			return Dam{}, errgo.Wrap(err, "nsfw_word")
		}
	}

	if c.DisableWords != "" {
		cc.disableWord, err = regexp.Compile(c.DisableWords)
		if err != nil {
			return Dam{}, errgo.Wrap(err, "disable_words")
		}
	}

	if c.BannedDomain != "" {
		cc.bannedDomain, err = regexp.Compile(c.BannedDomain)
		if err != nil {
			return Dam{}, errgo.Wrap(err, "banned_domain")
		}
	}

	return cc, nil
}

func (d Dam) NeedReview(text string) bool {
	if text == "" {
		return false
	}
	if d.disableWord == nil {
		return false
	}

	text = strings.ToLower(text)

	return d.disableWord.MatchString(text)
}

func (d Dam) CensoredWords(text string) bool {
	if d.disableWord != nil && d.disableWord.MatchString(text) {
		return true
	}

	if d.bannedDomain != nil && d.bannedDomain.MatchString(text) {
		return true
	}

	return false
}

func AllPrintableChar(text string) bool {
	for _, c := range text {
		switch c {
		case '\n', '\t':
			continue
		}

		if !unicode.IsPrint(c) {
			return false
		}
	}

	return true
}
