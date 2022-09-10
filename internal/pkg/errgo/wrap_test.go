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

//nolint:goerr113
package errgo_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/errgo"
)

func TestWrap(t *testing.T) {
	t.Parallel()
	err := errors.New("raw")
	require.Equal(t, "wrap: raw", errgo.Wrap(err, "wrap").Error())
	require.Equal(t, "e: wrap: raw", errgo.Wrap(errgo.Wrap(err, "wrap"), "e").Error())
}

func TestStackTrace(t *testing.T) {
	t.Parallel()

	err := errgo.Wrap(errors.New("a error"), "m")
	s := fmt.Sprintf("%+v", err)
	require.Regexp(t,
		regexp.MustCompile("^m: a error\nerror stack:\n.*errgo_test\\.TestStackTrace\n.*wrap_test.go:\\d+\n.*"),
		s)
}

func TestErrorIs(t *testing.T) {
	t.Parallel()

	e := errors.New("expected")

	err := errgo.Wrap(e, "ctx")
	require.True(t, errors.Is(err, e))

	err = errgo.MsgNoTrace(e, "ctx")
	require.True(t, errors.Is(err, e))

	err = errgo.Msg(e, "ctx")
	require.True(t, errors.Is(err, e))
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	e := errors.New("expected")
	err := errgo.Wrap(e, "ctx")

	b, jerr := json.Marshal(err)
	require.NoError(t, jerr)

	var m struct {
		Error string   `json:"error"`
		Stack []string `json:"stack"`
	}

	require.NoError(t, json.Unmarshal(b, &m))

	require.Equal(t, "ctx: expected", m.Error)
	require.NotZero(t, len(m.Stack), "stack should not be zero")
}
