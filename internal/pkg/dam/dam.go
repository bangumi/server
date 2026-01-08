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
	"unicode"

	"github.com/trim21/errgo"

	"github.com/bangumi/server/config"
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
		cc.nsfwWord, err = regexp.Compile("(?i)" + c.NsfwWord)
		if err != nil {
			return Dam{}, errgo.Wrap(err, "nsfw_word")
		}
	}

	if c.DisableWords != "" {
		cc.disableWord, err = regexp.Compile("(?i)" + c.DisableWords)
		if err != nil {
			return Dam{}, errgo.Wrap(err, "disable_words")
		}
	}

	if c.BannedDomain != "" {
		cc.bannedDomain, err = regexp.Compile("(?i)" + c.BannedDomain)
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
		case '\n', '\t', '\r':
			continue
		}

		if !unicode.IsPrint(c) {
			return false
		}
	}

	return true
}

var ZeroWidthPattern = regexp.MustCompile(`[^\t\r\n\p{L}\p{M}\p{N}\p{P}\p{S}\p{Z}]`)
var ExtraSpacePattern = regexp.MustCompile("[\u3000 ]")

func ValidateTag(t string) bool {
	if len(t) == 0 {
		return false
	}

	if !AllPrintableChar(t) {
		return false
	}

	if ZeroWidthPattern.MatchString(t) {
		return false
	}

	if ExtraSpacePattern.MatchString(t) {
		return false
	}

	return true
}
