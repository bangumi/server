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

//nolint:tagliatelle
package hcaptcha

import (
	"context"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/pkg/errgo"
)

const VerifyURL = "https://hcaptcha.com/siteverify"

type Manager struct {
	http   *resty.Client
	secret string
}

func New(cfg config.AppConfig, http *resty.Client) Manager {
	return Manager{secret: cfg.HCaptchaSecretKey, http: http}
}

func (m Manager) Verify(ctx context.Context, response string) (bool, error) {
	resp, err := m.http.R().SetFormData(map[string]string{
		"response": response,
		"secret":   m.secret,
	}).SetContext(ctx).Post(VerifyURL)

	if err != nil {
		return false, errgo.Wrap(err, "http request")
	}

	var d hCaptcha
	if err := sonic.Unmarshal(resp.Body(), &d); err != nil {
		return false, errgo.Wrap(err, "sonic.Unmarshal")
	}

	return d.Success, nil
}

type hCaptcha struct {
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
	ScoreReason []any     `json:"score_reason"`
	Score       float64   `json:"score"`
	Credit      bool      `json:"credit"`
	Success     bool      `json:"success"`
}
