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

package hcaptcha_test

import (
	"context"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/internal/web/captcha/hcaptcha"
)

func TestManager_Verify(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvExternalHTTP)
	// testing key, checkout https://docs.hcaptcha.com/#integration-testing-test-keys
	manager := hcaptcha.New(config.AppConfig{
		HCaptchaSecretKey: "0x0000000000000000000000000000000000000000",
	}, resty.New())

	ok, err := manager.Verify(context.Background(), "10000000-aaaa-bbbb-cccc-000000000001")
	if err != nil {
		t.Fatal("unexpected hCaptcha error, you may need to set a proxy with `HTTPS_PROXY`")
	}

	require.True(t, ok)
}

func TestManager_Verify_fail(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvExternalHTTP)
	// testing key, checkout https://docs.hcaptcha.com/#integration-testing-test-keys
	manager := hcaptcha.New(config.AppConfig{
		HCaptchaSecretKey: "0x0000000000000000000000000000000000000000",
	}, resty.New())

	ok, err := manager.Verify(context.Background(), "10000000-aaaa-bbbb-cccc-000000000002")
	if err != nil {
		t.Fatal("unexpected hCaptcha error, you may need to set a proxy with `HTTPS_PROXY`")
	}

	require.False(t, ok)
}
