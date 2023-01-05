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

package ctrl

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/bangumi/server/ctrl/internal/cachekey"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/res"
)

func (ctl Ctrl) GetIndexWithCache(c context.Context, id uint32) (res.Index, bool, error) {
	var key = cachekey.Index(id)

	var r res.Index
	ok, err := ctl.cache.Get(c, key, &r)
	if err != nil {
		return r, ok, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		return r, ok, nil
	}

	i, err := ctl.index.Get(c, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.Index{}, false, nil
		}

		return res.Index{}, false, errgo.Wrap(err, "Index.Get")
	}

	u, err := ctl.user.GetByID(c, i.CreatorID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			ctl.log.Error("index missing creator", zap.Uint32("index_id", id), i.CreatorID.Zap())
		}
		return res.Index{}, false, errgo.Wrap(err, "failed to get creator: user.GetByID")
	}

	r = res.IndexModelToResponse(&i, u)

	if e := ctl.cache.Set(c, key, r, time.Hour); e != nil {
		ctl.log.Error("can't set response to cache", zap.Error(e))
	}

	return r, true, nil
}
