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

package domain

import (
	"context"
	"time"
)

// AuthRepo presents an authorization.
type AuthRepo interface {
	// GetByToken return an authorized user by a valid access token.
	GetByToken(ctx context.Context, token string) (Auth, error)
}

// Auth is the basic authorization represent a user.
type Auth struct {
	RegTime time.Time
	ID      uint32
}

const nsfwThreshold = -time.Hour * 24 * 60

// AllowNSFW return if current user is allowed to see NSFW resource.
func (u Auth) AllowNSFW() bool {
	if u.ID == 0 {
		return false
	}

	return u.RegTime.Add(nsfwThreshold).Before(time.Now())
}
