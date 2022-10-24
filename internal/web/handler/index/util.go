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

package index

import (
	"context"

	"github.com/bangumi/server/internal/compat"
	"github.com/bangumi/server/internal/dam"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/web/handler/internal/cachekey"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/pkg/wiki"
)

func indexSubjectToResp(s domain.IndexSubject) res.IndexSubjectV0 {
	return res.IndexSubjectV0{
		AddedAt: s.AddedAt,
		Date:    null.NilString(s.Subject.Date),
		Image:   res.SubjectImage(s.Subject.Image),
		Name:    s.Subject.Name,
		NameCN:  s.Subject.NameCN,
		Comment: s.Comment,
		Infobox: compat.V0Wiki(wiki.ParseOmitError(s.Subject.Infobox).NonZero()),
		ID:      s.Subject.ID,
		TypeID:  s.Subject.TypeID,
	}
}

func (h Handler) invalidateIndexCache(ctx context.Context, id uint32) {
	// ignore error
	_ = h.cache.Del(ctx, cachekey.Index(id))
}

func (h Handler) ensureValidStrings(strings ...string) error {
	for _, str := range strings {
		if !dam.AllPrintableChar(str) {
			return res.BadRequest("invalid string")
		}
	}
	return nil
}
