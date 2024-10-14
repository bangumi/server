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

package tag

import (
	"context"

	"github.com/bangumi/server/internal/model"
)

// CatSubject 条目tag.
const CatSubject = 0

// CatMeta 官方tag.
const CatMeta = 3

type Tag struct {
	Name  string
	Count uint
	// TotalCount count for all tags including all subject
	TotalCount uint
}

type CachedRepo interface {
	read
}

type Repo interface {
	read
}

type read interface {
	Get(ctx context.Context, id model.SubjectID) ([]Tag, error)
	GetByIDs(ctx context.Context, ids []model.SubjectID) (map[model.SubjectID][]Tag, error)
}
