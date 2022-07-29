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

package null_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/null"
)

func TestValidateNull(t *testing.T) {
	t.Parallel()

	v := validator.New()
	v.RegisterCustomTypeFunc(null.ValidateNull, null.Int8{})

	var s = struct {
		I null.Int8 `json:"i" validate:"required"`
	}{
		I: null.Int8{Set: false, Null: true},
	}

	err := v.Struct(s)

	require.Error(t, err)
}

func TestValidateNull_no_error(t *testing.T) {
	t.Parallel()

	v := validator.New()
	v.RegisterCustomTypeFunc(null.ValidateNull, null.Int8{})

	var s = struct {
		I null.Int8 `json:"i" validate:"required"`
	}{
		I: null.Int8{Value: 5, Set: true, Null: false},
	}

	err := v.Struct(s)

	require.NoError(t, err)
}

func TestValidateNull_value(t *testing.T) {
	t.Parallel()

	v := validator.New()
	v.RegisterCustomTypeFunc(null.ValidateNull, null.Int8{})

	var s = struct {
		I null.Int8 `json:"i" validate:"required,lte=5"`
	}{
		I: null.Int8{Value: 6, Set: true, Null: false},
	}

	err := v.Struct(s)
	require.Error(t, err)
}
