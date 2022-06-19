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

package req_test

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/req"
)

/*
go test ./internal/web/req -run '^TestPutSubjectCollection_validation$'
*/
func TestPutSubjectCollection_validation(t *testing.T) {
	t.Parallel()
	v := validator.New()
	tests := []req.PutSubjectCollection{
		{Rate: 1, Type: model.CollectionWish},
	}

	for _, s := range tests {
		t.Run(fmt.Sprintf("type=%d,rate=%d", s.Type, s.Rate), func(t *testing.T) {
			t.Parallel()
			err := v.Struct(s)
			require.NoError(t, err)
		})
	}
}

/*
go test ./internal/web/req -run '^TestPutSubjectCollection_validation_error$'
*/
func TestPutSubjectCollection_validation_error(t *testing.T) {
	t.Parallel()
	v := validator.New()
	tests := []req.PutSubjectCollection{
		{Rate: 11, Type: model.CollectionWish},
		{Rate: 0, Type: model.CollectionWish},
		{Rate: 5, Type: 6},
		{Rate: 5, Type: 0},
	}

	for _, s := range tests {
		t.Run(fmt.Sprintf("type=%d,rate=%d", s.Type, s.Rate), func(t *testing.T) {
			t.Parallel()
			err := v.Struct(s)
			require.Error(t, err)
		})
	}
}
