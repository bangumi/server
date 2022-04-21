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

package hcaptcha

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/web/captcha"
)

const VerifyURL = "https://hcaptcha.com/siteverify"

type manager struct {
	http   *resty.Client
	secret string
}

func New(cfg config.AppConfig, http *resty.Client) captcha.Manager {
	return manager{secret: cfg.HCaptchaSecretKey, http: http}
}

func (m manager) Verify(ctx context.Context, response string) (bool, error) {
	resp, err := m.http.R().SetFormData(map[string]string{
		"response": response,
		"secret":   m.secret,
	}).SetContext(ctx).Post(VerifyURL)

	if err != nil {
		return false, errgo.Wrap(err, "http request")
	}

	var d hCaptcha
	if err := json.Unmarshal(resp.Body(), &d); err != nil {
		return false, errgo.Wrap(err, "json.Unmarshal")
	}

	return true, nil
}

type hCaptcha struct {
	ErrorCodes  []string      `json:"error-codes"` //nolint:tagliatelle
	ScoreReason []interface{} `json:"score_reason"`
	ChallengeTS time.Time     `json:"challenge_ts"`
	Hostname    string        `json:"hostname"`
	Score       float64       `json:"score"`
	Credit      bool          `json:"credit"`
	Success     bool          `json:"success"`
}
