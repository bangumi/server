// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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

package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/auth"
	"github.com/bangumi/server/domain"
)

func getService() domain.AuthService {
	return auth.NewService(nil, zap.NewNop())
}

func TestService_ComparePassword(t *testing.T) {
	t.Parallel()
	s := getService()
	var hashed = []byte("$2a$12$GA5Pr9GhsyLJcSPoTpYBY.JqTzYZb2nfgSeZ1EK38bfgk/Rykkvuq")
	var input = "lovemeplease"

	eq, err := s.ComparePassword(hashed, input)
	require.NoError(t, err)
	require.True(t, eq)
}
