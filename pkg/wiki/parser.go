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

package wiki

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrWikiSyntax         = errors.New("invalid wiki syntax")
	ErrGlobalPrefix       = fmt.Errorf("%w: missing prefix '{{Infobox' at the start", ErrWikiSyntax)
	ErrGlobalSuffix       = fmt.Errorf("%w: missing '}}' at the end", ErrWikiSyntax)
	ErrArrayNoClose       = fmt.Errorf("%w: array should be closed by '}'", ErrWikiSyntax)
	ErrArrayItemWrapped   = fmt.Errorf("%w: array item should be wrapped by '[]'", ErrWikiSyntax)
	ErrExpectingNewField  = fmt.Errorf("%w: missing '|' to start a new field", ErrWikiSyntax)
	ErrExpectingSignEqual = fmt.Errorf("%w: missing '=' to separate field name and value", ErrWikiSyntax)
)

// ParseOmitError try to parse a string as wiki, omitting error.
func ParseOmitError(s string) Wiki {
	w, _ := Parse(s)

	return w
}

const prefix = "{{Infobox"
const suffix = "}}"

//nolint:funlen,gocognit,gocyclo
func Parse(s string) (Wiki, error) {
	var w = Wiki{}
	s, lineOffset := processInput(s)
	if s == "" {
		return w, nil
	}

	if !strings.HasPrefix(s, prefix) {
		return Wiki{}, ErrGlobalPrefix
	}

	eolCount := strings.Count(s, "\n")

	if !strings.HasSuffix(s, suffix) {
		return Wiki{}, ErrGlobalSuffix
	}

	w.Type = readType(s)
	w.Fields = make([]Field, 0) // make zero value in json '[]', no alloc with cap 0

	if eolCount <= 1 {
		return w, nil
	}

	w.Fields = make([]Field, 0, eolCount-1)
	// pre-alloc for all items.
	var itemContainer = make([]Item, 0, eolCount-2)
	var lastCut = 0
	var currentCut = 0

	// loop state
	var inArray = false
	var currentField Field

	// variable to loop line
	var firstEOL = strings.IndexByte(s, '\n') // skip first line
	var secondLastEOL = 0
	var lastEOL = firstEOL + 1
	var lino = lineOffset - 1 // current line number
	var offset int
	var line string
	for {
		// fast iter lines without alloc
		offset = strings.IndexByte(s[lastEOL:], '\n')
		if offset != -1 {
			line = s[lastEOL : lastEOL+offset]
			secondLastEOL = lastEOL
			lastEOL = lastEOL + offset + 1
			lino++
		} else {
			// can't find next line
			if inArray {
				// array should be close have read all contents
				return Wiki{}, wrapError(ErrArrayNoClose, lino+1, s[secondLastEOL:lastEOL])
			}

			break
		}

		// now handle line content
		line = trimSpace(line)
		if line == "" {
			continue
		}

		if line[0] == '|' {
			// new field
			currentField = Field{}
			if inArray {
				return Wiki{}, wrapError(ErrArrayNoClose, lino, line)
			}

			key, value, err := readStartLine(trimLeftSpace(line[1:])) // read "key = value"
			if err != nil {
				return Wiki{}, wrapError(err, lino, line)
			}

			switch value {
			case "":
				w.Fields = append(w.Fields, Field{Key: key, Null: true})

				continue

			case "{":
				inArray = true
				currentField.Key = key
				currentField.Array = true

				continue
			}

			w.Fields = append(w.Fields, Field{Key: key, Value: value})

			continue
		}

		if inArray {
			if line == "}" { // close array
				inArray = false
				currentField.Values = itemContainer[lastCut:currentCut]
				lastCut = currentCut
				w.Fields = append(w.Fields, currentField)

				continue
			}
			// array item
			key, value, err := readArrayItem(line)
			if err != nil {
				return Wiki{}, wrapError(err, lino, line)
			}
			itemContainer = append(itemContainer, Item{
				Key:   key,
				Value: value,
			})
			currentCut++
		}

		if !inArray {
			return Wiki{}, wrapError(ErrExpectingNewField, lino, line)
		}
	}

	return w, nil
}

func readType(s string) string {
	i := strings.IndexByte(s, '\n')
	if i == -1 {
		i = strings.IndexByte(s, '}') // {{Infobox Crt}}
	}

	return trimSpace(s[len(prefix):i])
}

// read whole line as an array item, spaces are trimmed.
//   readArrayItem("[简体中文名|鲁鲁修]") => "简体中文名", "鲁鲁修", nil
//   readArrayItem("[简体中文名|]") => "简体中文名", "", nil
//   readArrayItem("[鲁鲁修]") => "", "鲁鲁修", nil
func readArrayItem(line string) (string, string, error) {
	if line[0] != '[' || line[len(line)-1] != ']' {
		return "", "", ErrArrayItemWrapped
	}

	content := line[1 : len(line)-1]

	i := strings.IndexByte(content, '|')
	if i == -1 {
		return "", trimSpace(content), nil
	}

	return trimSpace(content[:i]), trimSpace(content[i+1:]), nil
}

// read line without leading '|' as key value pair, spaces are trimmed.
//   readStartLine("播放日期 = 2017年4月16日") => 播放日期, 2017年4月16日, nil
//   readStartLine("播放日期 = ") => 播放日期, "", nil
func readStartLine(line string) (string, string, error) {
	i := strings.IndexByte(line, '=')
	if i == -1 {
		return "", "", ErrExpectingSignEqual
	}

	return trimRightSpace(line[:i]), trimLeftSpace(line[i+1:]), nil
}
