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

package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/domain"
)

func getService() domain.AuthService {
	return NewService(nil)
}

func testService_ComparePassword(t *testing.T) {
	t.Parallel()
	s := getService()
	// TODO: 用树洞号的帐号密码测试
	var hashed []byte
	var input string

	eq, err := s.ComparePassword(hashed, input)
	require.NoError(t, err)
	require.True(t, eq)
}
