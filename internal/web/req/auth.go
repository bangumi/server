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

package req

type UserLogin struct {
	Email            string `json:"email" validate:"required,email"`
	Password         string `json:"password" validate:"required"`
	HCaptchaResponse string `json:"h-captcha-response" validate:"required"` //nolint:tagliatelle
}

type CreatePersonalAccessToken struct {
	Name         string `json:"name"`
	DurationDays uint   `json:"duration_days" validate:"required,lte=365" validateName:"有效期"`
}

type DeletePersonalAccessToken struct {
	ID uint32 `json:"id" validate:"required"`
}
