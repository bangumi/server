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

func TestValidate(t *testing.T) {
	t.Parallel()

	v := validator.New()
	v.RegisterCustomTypeFunc(null.Validator, null.AllTypes()...)

	t.Run("value", func(t *testing.T) {
		t.Parallel()
		var field = null.NewUint8(5)
		errs := v.Var(field, "lt=3")
		require.NotNil(t, errs)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		var field = null.Uint8{
			Uint8: 5,
			Set:   false,
			Null:  true,
		}
		errs := v.Var(field, "lt=3,omitempty")
		require.NotNil(t, errs)
	})
}
