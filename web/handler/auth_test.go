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

package handler_test

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/mocks"
	"github.com/bangumi/server/model"
)

func TestHandler_PrivateLogin(t *testing.T) {
	t.Parallel()

	m := &mocks.SubjectRepo{}
	m.EXPECT().Get(mock.Anything, uint32(7)).Return(model.Subject{NSFW: true}, nil)

	// 原本的代码库就是这么储存密码的...
	p := md5.Sum([]byte("p")) //nolint:gosec
	var h = make([]byte, 32)
	hex.Encode(h, p[:])
	passwordInDB, err := bcrypt.GenerateFromPassword(h, bcrypt.DefaultCost)
	require.NoError(t, err)

	mockAuth := &mocks.AuthRepo{}
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{}, nil)
	mockAuth.EXPECT().GetByEmail(mock.Anything, "user email").Return(domain.Auth{}, passwordInDB, nil)

	mockCaptcha := &mocks.CaptchaManager{}
	mockCaptcha.EXPECT().Verify(mock.Anything, "req").Return(true, nil)
	// defer mockCaptcha.AssertExpectations(t)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo:       mockAuth,
			SubjectRepo:    m,
			CaptchaManager: mockCaptcha,
		},
	)

	resp := test.New(t).Post("/p/login").JSON(fiber.Map{
		"email":              "user email",
		"password":           "p",
		"h-captcha-response": "req",
	}).Header(fiber.HeaderUserAgent, "fiber test client").
		Execute(app, -1)

	require.Equal(t, fiber.StatusOK, resp.StatusCode, resp.BodyString())

	_, ok := resp.Header[fiber.HeaderSetCookie]
	require.True(t, ok, "response should set cookies")
	require.Equal(t, http.StatusOK, resp.StatusCode, "200 for login")
}
