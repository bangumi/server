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

package syntax_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bangumi/server/internal/search/syntax"
)

func TestParse(t *testing.T) {
	t.Parallel()
	const testString = `one -two tag:"three four" -ff:1 a:>2 b:<3`
	var result, err = syntax.Parse(testString)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, syntax.Result{
		Keyword: []string{"one", "-two"},
		Filter: map[string][]string{
			"tag": {"three four"},
			"-ff": {"1"},
			"a":   {">2"},
			"b":   {"<3"},
		},
	}, result)
}

func TestTokenizerBroken(t *testing.T) {
	t.Parallel()
	const testString = `one -two tag:'three four' "-ff:1"""'`

	_, err := syntax.Parse(testString)
	assert.NotNil(t, err)
	assert.Regexp(t, regexp.MustCompile("EOF.*closing quote"), err.Error())
}
