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

package wiki_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/bangumi/server/pkg/wiki"
)

const testRoot = "./testdata/wiki-syntax-spec/tests/"

type Result struct {
	Type string  `yaml:"type"`
	Data []Field `yaml:"data"`
}

func (r Result) Wiki() wiki.Wiki {
	fields := make([]wiki.Field, len(r.Data))
	for i, datum := range r.Data {
		fields[i] = datum.Wiki()
	}
	return wiki.Wiki{Type: r.Type, Fields: fields}
}

type Field struct {
	Key    string `yaml:"key"`
	Value  string `yaml:"value"`
	Values []Item `yaml:"values"`
	Array  bool   `yaml:"array"`
}

func (i Field) Wiki() wiki.Field {
	var values []wiki.Item
	if i.Array {
		values = make([]wiki.Item, len(i.Values))
		for i, datum := range i.Values {
			values[i] = datum.Wiki()
		}
	}
	return wiki.Field{
		Key:    i.Key,
		Value:  i.Value,
		Values: values,
		Array:  i.Array,
		Null:   len(i.Values) == 0 && i.Value == "",
	}
}

type Item struct {
	K string `yaml:"k"`
	V string `yaml:"v"`
}

func (i Item) Wiki() wiki.Item {
	return wiki.Item{
		Key:   i.K,
		Value: i.V,
	}
}

func TestAgainstInvalidSpec(t *testing.T) {
	t.Parallel()
	checkSubmodule(t)
	var caseRoot = filepath.Join(testRoot, "invalid")
	files, err := ioutil.ReadDir(caseRoot)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		// name := file.Name()
		file := file
		t.Run(file.Name(), func(t *testing.T) {
			t.Parallel()
			raw, err := os.ReadFile(filepath.Join(caseRoot, file.Name()))
			require.NoError(t, err)
			_, err = wiki.Parse(string(raw))
			require.NotNil(t, err, "expecting reporting error")
		})
	}
}

func TestAgainstValidSpec(t *testing.T) {
	t.Parallel()
	checkSubmodule(t)
	var caseRoot = filepath.Join(testRoot, "valid")
	files, err := ioutil.ReadDir(caseRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files { //nolint:paralleltest
		if strings.HasSuffix(file.Name(), ".wiki") {
			name := strings.TrimSuffix(file.Name(), ".wiki")
			t.Run(name, testCase(caseRoot, name))
		}
	}
}

func testCase(root, name string) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		raw, err := os.ReadFile(filepath.Join(root, name+".wiki"))
		require.NoError(t, err)

		yamlRaw, err := os.ReadFile(filepath.Join(root, name+".yaml"))
		require.NoError(t, err)

		expected := Result{}
		require.NoError(t, yaml.Unmarshal(yamlRaw, &expected))

		result, err := wiki.Parse(string(raw))
		require.NoError(t, err)

		require.Equal(t, expected.Wiki(), result)
	}
}

func checkSubmodule(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(testRoot); err != nil && os.IsNotExist(err) {
		t.Fatal("test data missing, do you forget to init git submodules?" +
			"Try `git submodule update --init --recursive`")
	}
}
