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
	"bytes"
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
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

	body, err := json.Marshal(fiber.Map{
		"email":              "user email",
		"password":           "p",
		"h-captcha-response": "req",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/p/login", bytes.NewBuffer(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderUserAgent, "fiber test client")

	resp, err := app.Test(req, -1) // bcrypt 比较哈希太慢了
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(respBody))

	_, ok := resp.Header[fiber.HeaderSetCookie]
	require.True(t, ok, "response should set cookies")
	require.Equal(t, http.StatusOK, resp.StatusCode, "200 for login")
}
