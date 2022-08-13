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
	"strings"
	"testing"
)

func TestClassifier(t *testing.T) {
	t.Parallel()
	classifier := newDefaultClassifier()
	tests := map[rune]runeTokenClass{
		' ':  spaceRuneClass,
		'"':  doubleQuoteRuneClass,
		'\'': singleQuoteRuneClass,
		':':  filterSplitClass,
	}
	for runeChar, want := range tests {
		got := classifier.ClassifyRune(runeChar)
		if got != want {
			t.Errorf("ClassifyRune(%v) -> %v. Want: %v", runeChar, got, want)
		}
	}
}

func TestTokenizer(t *testing.T) {
	t.Parallel()
	const testString = `one -two tag:"three four" -ff:1 a:>2 b:<3`
	testInput, expectedTokens := strings.NewReader(testString), []*Token{
		{"one", KeywordToken},
		{"-two", KeywordToken},
		{"tag", FilterKeyToken},
		{"three four", FilterValueToken},

		{"-ff", FilterKeyToken},
		{"1", FilterValueToken},

		{"a", FilterKeyToken},
		{">2", FilterValueToken},

		{"b", FilterKeyToken},
		{"<3", FilterValueToken},
	}

	tokenizer := NewTokenizer(testInput)
	for i, want := range expectedTokens {
		got, err := tokenizer.Next()
		if err != nil {
			t.Error(err)
		}
		if !got.Equal(want) {
			t.Errorf("Tokenizer.Next()[%v] of %q -> %v. Want: %v", i, testString, got, want)
		}
	}
}

func TestTokenizer2(t *testing.T) {
	t.Parallel()
	const testString = `one -two tag:'three four' "-ff:1" a:>2 b:<3`
	testInput, expectedTokens := strings.NewReader(testString), []*Token{
		{"one", KeywordToken},
		{"-two", KeywordToken},
		{"tag", FilterKeyToken},
		{"three four", FilterValueToken},

		{"-ff:1", KeywordToken},

		{"a", FilterKeyToken},
		{">2", FilterValueToken},

		{"b", FilterKeyToken},
		{"<3", FilterValueToken},
	}

	tokenizer := NewTokenizer(testInput)
	for i, want := range expectedTokens {
		got, err := tokenizer.Next()
		if err != nil {
			t.Error(err)
		}
		if !got.Equal(want) {
			t.Errorf("Tokenizer.Next()[%v] of %q -> %v. Want: %v", i, testString, got, want)
		}
	}
}
