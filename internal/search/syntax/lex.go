/*
Copyright 2021 trim21
Copyright 2012 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package syntax

import (
	"bufio"
	"io"

	"github.com/pkg/errors"
)

var ErrParse = errors.New("failed to parse")

// TokenType is a top-level Token classification: A word, space, comment, unknown.
type TokenType int

// runeTokenClass is the type of a UTF-8 character classification: A quote, space, escape.
type runeTokenClass int

// the internal state used by the lexer state machine.
type lexerState int

// Token is a (type, value) pair representing a lexographical Token.
type Token struct {
	value     string
	tokenType TokenType
}

// Equal reports whether tokens a, and b, are equal.
// Two tokens are equal if both their types and values are equal. A nil Token can
// never be equal to another Token.
func (a *Token) Equal(b *Token) bool {
	if a == nil || b == nil {
		return false
	}
	if a.tokenType != b.tokenType {
		return false
	}

	return a.value == b.value
}

// Named classes of UTF-8 runes.
const (
	spaceRunes       = " \t"
	doubleQuoteRunes = `"`
	singleQuoteRunes = "'"
	filterSplitRunes = ":"
)

// Classes of rune Token.
const (
	spaceRuneClass runeTokenClass = iota + 1
	doubleQuoteRuneClass
	singleQuoteRuneClass
	filterSplitClass
	eofRuneClass
)

// Classes of lexographic Token.
const (
	KeywordToken TokenType = iota + 1
	FilterKeyToken
	FilterValueToken
)

// Lexer state machine states.
const (
	nilState           lexerState = iota // zero value of state
	startState                           // no runes have been seen
	inWordState                          // processing regular runes in a word
	inFilterState                        // processing regular runes in a word
	doubleQuotingState                   // we are within a quoted string that supports escaping ("...")
	singleQuotingState                   // we are within a string that does not support escaping ('...')
)

// tokenClassifier is used for classifying rune characters.
type tokenClassifier map[rune]runeTokenClass

func (t tokenClassifier) addRuneClass(runes string, tokenType runeTokenClass) {
	for _, runeChar := range runes {
		t[runeChar] = tokenType
	}
}

// newDefaultClassifier creates a new classifier for ASCII characters.
func newDefaultClassifier() tokenClassifier {
	// initialize it out current runes count
	var t = make(tokenClassifier, 5) //nolint:gomnd

	t.addRuneClass(spaceRunes, spaceRuneClass)
	t.addRuneClass(doubleQuoteRunes, doubleQuoteRuneClass)
	t.addRuneClass(singleQuoteRunes, singleQuoteRuneClass)
	t.addRuneClass(filterSplitRunes, filterSplitClass)

	return t
}

// ClassifyRune classifieds a rune.
func (t tokenClassifier) ClassifyRune(runeVal rune) runeTokenClass {
	return t[runeVal]
}

// Tokenizer turns an input stream into a sequence of typed tokens.
type Tokenizer struct {
	classifier tokenClassifier
	input      bufio.Reader
	state      lexerState
}

// NewTokenizer creates a new tokenizer from an input stream.
func NewTokenizer(r io.Reader) *Tokenizer {
	input := bufio.NewReader(r)
	classifier := newDefaultClassifier()

	return &Tokenizer{
		input:      *input,
		classifier: classifier,
	}
}

// scanStream scans the stream for the next Token using the internal state machine.
//nolint:gocyclo,funlen,cyclop
func (t *Tokenizer) scanStream() (*Token, error) {
	state := startState
	var tokenType TokenType
	var value []rune
	var nextRune rune
	var nextRuneType runeTokenClass
	var err error

	for {
		nextRune, _, err = t.input.ReadRune()
		nextRuneType = t.classifier.ClassifyRune(nextRune)

		if err == io.EOF { // nolint: errorlint
			nextRuneType = eofRuneClass
			err = nil
		} else if err != nil {
			return nil, err
		}

		switch state {
		case startState: // no runes read yet
			{
				switch nextRuneType {
				case eofRuneClass:
					{
						return nil, io.EOF
					}
				case spaceRuneClass:
					{
					}
				case doubleQuoteRuneClass:
					{
						tokenType = KeywordToken
						state = doubleQuotingState
					}
				case singleQuoteRuneClass:
					{
						tokenType = KeywordToken
						state = singleQuotingState
					}
				default:
					{
						tokenType = KeywordToken
						value = append(value, nextRune)
						state = inWordState
					}
				}
			}
		case inWordState: // in a regular word
			{
				switch nextRuneType {
				case eofRuneClass:
					{
						if t.state == inFilterState {
							t.state = nilState

							return &Token{tokenType: FilterValueToken, value: string(value)}, nil
						}

						return &Token{tokenType: tokenType, value: string(value)}, err
					}
				case spaceRuneClass:
					if t.state == inFilterState {
						t.state = nilState

						return &Token{tokenType: FilterValueToken, value: string(value)}, nil
					}

					{
						token := &Token{
							tokenType: tokenType,
							value:     string(value)}

						return token, err
					}
				case filterSplitClass:
					{
						t.state = inFilterState
						token := &Token{
							tokenType: FilterKeyToken,
							value:     string(value)}

						return token, err
					}
				case doubleQuoteRuneClass:
					{
						state = doubleQuotingState
					}
				case singleQuoteRuneClass:
					{
						state = singleQuotingState
					}
				default:
					{
						value = append(value, nextRune)
					}
				}
			}
		case doubleQuotingState: // in escaping double quotes
			{
				switch nextRuneType {
				case eofRuneClass:
					{
						return &Token{tokenType: tokenType, value: string(value)},
							errors.WithMessagef(ErrParse, "EOF found when expecting closing quote")
					}
				case doubleQuoteRuneClass:
					{
						state = inWordState
					}
				default:
					{
						value = append(value, nextRune)
					}
				}
			}
		case singleQuotingState: // in non-escaping single quotes
			{
				switch nextRuneType {
				case eofRuneClass:
					{
						return &Token{tokenType: tokenType, value: string(value)},
							errors.WithMessage(ErrParse, "EOF found when expecting closing quote")
					}
				case singleQuoteRuneClass:
					{
						state = inWordState
					}
				default:
					{
						value = append(value, nextRune)
					}
				}
			}
		default:
			{
				return nil, errors.WithMessagef(ErrParse, "Unexpected state: %v", state)
			}
		}
	}
}

// Next returns the next Token in the stream.
func (t *Tokenizer) Next() (*Token, error) {
	return t.scanStream()
}
