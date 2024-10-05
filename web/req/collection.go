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

import (
	"strings"
	"unicode/utf8"

	"github.com/samber/lo"

	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/pkg/dam"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/web/res"
)

type UpdateEpisodeCollection struct {
	Type uint8 `json:"type"`
}

type SubjectEpisodeCollectionPatch struct {
	Comment   null.String                             `json:"comment"`
	Tags      []string                                `json:"tags"`
	VolStatus null.Uint32                             `json:"vol_status" doc:"只能用于书籍条目"`
	EpStatus  null.Uint32                             `json:"ep_status" doc:"只能用于书籍条目"`
	Type      null.Null[collection.SubjectCollection] `json:"type"`
	Rate      null.Uint8                              `json:"rate"`
	Private   null.Bool                               `json:"private"`
}

func (v *SubjectEpisodeCollectionPatch) Validate() error {
	if v.Rate.Set {
		if v.Rate.Value > 10 {
			return res.BadRequest("rate overflow")
		}
	}

	if len(v.Tags) > 0 {
		if !lo.EveryBy(v.Tags, dam.AllPrintableChar) {
			return res.BadRequest("invisible character are included in tags")
		}

		if lo.ContainsBy(v.Tags, func(item string) bool {
			return len(item) == 0
		}) {
			return res.BadRequest("zero length tags are included in tags")
		}
	}

	if v.Comment.Set {
		if !dam.AllPrintableChar(v.Comment.Value) {
			return res.BadRequest("invisible character are included in comment")
		}

		v.Comment.Value = strings.TrimSpace(v.Comment.Value)
		if utf8.RuneCountInString(v.Comment.Value) > 380 {
			return res.BadRequest("comment too long, only allow less equal than 380 characters")
		}
	}

	return nil
}

type UpdateUserEpisodeCollection struct {
	Type collection.EpisodeCollection `json:"type"`
}
