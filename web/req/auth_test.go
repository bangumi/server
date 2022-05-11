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

package req_test

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/web/req"
)

func TestLoginPass(t *testing.T) {
	t.Parallel()
	var testCase = []req.UserLogin{
		{Email: "a@b.com", Password: "qa", HCaptchaResponse: "qe"},
		{Email: "abc@abc.com", Password: "qqqqqqqq", HCaptchaResponse: "q"},
	}
	validate := validator.New()
	for i, login := range testCase {
		login := login
		t.Run(fmt.Sprintf("success %d", i), func(t *testing.T) {
			t.Parallel()
			require.NoError(t, validate.Struct(login))
		})
	}
}

func TestLoginErr(t *testing.T) {
	t.Parallel()
	var testCase = []req.UserLogin{
		{Email: "b", Password: "qa", HCaptchaResponse: "qe"},
		{Email: "1", Password: "qqqqq", HCaptchaResponse: "q"},
		{Email: "abc@abc.com", Password: "", HCaptchaResponse: "q"},
		{Email: "abc@abc.com", Password: "q", HCaptchaResponse: ""},
	}
	validate := validator.New()
	for i, login := range testCase {
		login := login
		t.Run(fmt.Sprintf("fail %d", i), func(t *testing.T) {
			t.Parallel()
			require.Error(t, validate.Struct(login))
		})
	}
}
