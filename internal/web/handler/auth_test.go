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
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gookit/goutil/timex"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/internal/web/session"
)

func TestHandler_PrivateLogin(t *testing.T) {
	t.Parallel()

	// 原本的代码库就是这么储存密码的...
	p := md5.Sum([]byte("p")) //nolint:gosec
	var h = make([]byte, 32)
	hex.Encode(h, p[:])
	passwordInDB, err := bcrypt.GenerateFromPassword(h, bcrypt.DefaultCost)
	require.NoError(t, err)

	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByEmail(mock.Anything, "a@example.com").Return(domain.Auth{GroupID: 1}, passwordInDB, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, uint8(1)).Return(domain.Permission{}, nil)

	mockCaptcha := mocks.NewCaptchaManager(t)
	mockCaptcha.EXPECT().Verify(mock.Anything, "req").Return(true, nil)

	app := test.GetWebApp(t,
		test.Mock{
			AuthRepo:       mockAuth,
			CaptchaManager: mockCaptcha,
		},
	)

	resp := test.New(t).Post("/p/login").JSON(fiber.Map{
		"email":              "a@example.com",
		"password":           "p",
		"h-captcha-response": "req",
	}).Execute(app)

	require.Equal(t, fiber.StatusOK, resp.StatusCode, resp.BodyString())

	_, ok := resp.Header[fiber.HeaderSetCookie]
	require.True(t, ok, "response should set cookies")
	require.Equal(t, http.StatusOK, resp.StatusCode, "200 for login")
}

func TestHandler_PrivateLogin_content_type(t *testing.T) {
	t.Parallel()

	app := test.GetWebApp(t, test.Mock{})

	resp := test.New(t).Post("/p/login").Form("email", "abc@exmaple.com").Execute(app)

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, resp.BodyString())
}

func TestHandler_PrivateLogout(t *testing.T) {
	t.Parallel()

	mockCaptcha := mocks.NewSessionManager(t)
	mockCaptcha.EXPECT().Get(mock.Anything, "req").Return(session.Session{
		RegTime:   time.Now().Add(-timex.OneWeek),
		UserID:    1,
		ExpiredAt: time.Now().Unix() + timex.OneWeekSec,
	}, nil)
	mockCaptcha.EXPECT().Revoke(mock.Anything, "req").Return(nil)

	app := test.GetWebApp(t, test.Mock{SessionManager: mockCaptcha})

	resp := test.New(t).Post("/p/logout").Cookie(session.Key, "req").Execute(app)

	require.Equal(t, fiber.StatusNoContent, resp.StatusCode, resp.BodyString())

	var found bool
	for _, cookie := range resp.Cookies() {
		if cookie.Name == session.Key {
			found = true
			require.Equal(t, "", cookie.Value)
		}
	}
	require.True(t, found)
}
