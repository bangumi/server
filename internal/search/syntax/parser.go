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
	"io"
	"strings"

	"github.com/pkg/errors"
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
			if err == io.EOF { // nolint: errorlint
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
			return r, errors.WithMessagef(ErrParse, "unexpected token %+v", got)
		}
	}
}
