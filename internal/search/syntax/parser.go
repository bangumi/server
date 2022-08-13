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

/*
Package syntax implements a simple lexer and a parser to parse search syntax we are using

The basic use case uses the default ASCII lexer to split a string into sub-strings:

	  syntax.Parse(`one tag:"two three" four -tag:cc`) ->
			Struct{
				Keyword: []string{"one", "four"},
				Filter: map[string][]string{
					"tag":  []string{"two three"}
					"-tag": []string{"cc"}
				}
			}
*/
package syntax

import (
	"fmt"
	"io"
	"strings"

	"github.com/bangumi/server/internal/pkg/errgo"
)

type Result struct {
	Filter  map[string][]string
	Keyword []string
}

func Parse(s string) (Result, error) {
	var input = strings.NewReader(s)
	var r = Result{Filter: make(map[string][]string)}
	var f = ""

	tokenizer := NewTokenizer(input)
	for {
		got, err := tokenizer.Next()
		if err != nil {
			if err == io.EOF { //nolint:errorlint
				return r, nil
			}

			return r, err
		}

		switch got.tokenType {
		case KeywordToken:
			r.Keyword = append(r.Keyword, got.value)
		case FilterKeyToken:
			f = got.value
		case FilterValueToken:
			{
				r.Filter[f] = append(r.Filter[f], got.value)
				f = ""
			}
		default:
			return r, errgo.Msg(ErrParse, fmt.Sprintf("unexpected token %+v", got))
		}
	}
}
