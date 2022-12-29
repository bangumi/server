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

package session

import (
	"context"
	"errors"
	"time"

	"github.com/bangumi/server/internal/model"
)

const defaultRetry = 5

var errTooManyRetry = errors.New("too many reties with key conflict")

type Repo interface {
	Create(
		ctx context.Context, userID model.UserID, regTime time.Time, keyGen func() string,
	) (key string, s Session, err error)
	Get(ctx context.Context, key string) (Session, error)
	RevokeUser(ctx context.Context, userID model.UserID) (keys []string, err error)
	Revoke(ctx context.Context, key string) error
}
