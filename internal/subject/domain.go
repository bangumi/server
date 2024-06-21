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

package subject

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
)

type Filter struct {
	// if nsfw subject are allowed
	NSFW null.Bool
}

type BrowseFilter struct {
	NSFW     null.Bool
	Type     uint8
	Category null.Uint16
	Series   null.Bool
	Platform null.String
	Sort     null.String
	Year     null.Int32
	Month    null.Int8
}

func (f BrowseFilter) Hash() (string, error) {
	h := fnv.New64a()
	fields := []string{}
	fields = append(fields, fmt.Sprintf("type:%v", f.Type))
	if f.NSFW.Set {
		fields = append(fields, fmt.Sprintf("nsfw:%v", f.NSFW))
	}
	if f.Category.Set {
		fields = append(fields, fmt.Sprintf("category:%v", f.Category))
	}
	if f.Series.Set {
		fields = append(fields, fmt.Sprintf("series:%v", f.Series))
	}
	if f.Platform.Set {
		fields = append(fields, fmt.Sprintf("platform:%v", f.Platform))
	}
	if f.Sort.Set {
		fields = append(fields, fmt.Sprintf("sort:%v", f.Sort))
	}
	if f.Year.Set {
		fields = append(fields, fmt.Sprintf("year:%v", f.Year))
	}
	if f.Month.Set {
		fields = append(fields, fmt.Sprintf("month:%v", f.Month))
	}
	for _, field := range fields {
		if _, err := h.Write([]byte(field)); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", h.Sum64()), nil
}

type Repo interface {
	read
	post
}

type CachedRepo interface {
	read
}

type read interface {
	// Get return a repository model.
	Get(ctx context.Context, id model.SubjectID, filter Filter) (model.Subject, error)
	GetByIDs(ctx context.Context, ids []model.SubjectID, filter Filter) (map[model.SubjectID]model.Subject, error)

	Count(ctx context.Context, filter BrowseFilter) (int64, error)
	Browse(ctx context.Context, filter BrowseFilter, limit, offset int) ([]model.Subject, error)

	GetPersonRelated(ctx context.Context, personID model.PersonID) ([]domain.SubjectPersonRelation, error)
	GetCharacterRelated(ctx context.Context, characterID model.CharacterID) ([]domain.SubjectCharacterRelation, error)
	GetSubjectRelated(ctx context.Context, subjectID model.SubjectID) ([]domain.SubjectInternalRelation, error)

	GetActors(
		ctx context.Context, subjectID model.SubjectID, characterIDs []model.CharacterID,
	) (map[model.CharacterID][]model.PersonID, error)
}

type post interface {
	NewPost(ctx context.Context, post model.SubjectPost) error
	GetPaginatedPostsBySubjectID(ctx context.Context, id model.SubjectID, offset int, limit int) ([]model.SubjectPost, error)
	GetPostByID(ctx context.Context, id model.PostID) (model.SubjectPost, error)
	DeletePostByID(ctx context.Context, id model.PostID) error
	GetPaginatedRepliesByPostID(ctx context.Context, id model.PostID, offset int, limit int) ([]model.SubjectPost, error)
	GetPaginatedTopLevelPostsBySubjectID(ctx context.Context, id model.SubjectID, offset int, limit int) ([]model.SubjectPost, error)
}
