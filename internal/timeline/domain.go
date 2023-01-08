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

package timeline

import (
	"context"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
)

type Service interface {
	ChangeSubjectCollection(
		ctx context.Context,
		u model.UserID,
		sbj model.Subject,
		collect collection.SubjectCollection,
		comment string,
		rate uint8,
	) error

	ChangeEpisodeStatus(
		ctx context.Context,
		u auth.Auth,
		sbj model.Subject,
		episode episode.Episode,
	) error

	ChangeSubjectProgress(
		ctx context.Context,
		u model.UserID,
		sbj model.Subject,
		epsUpdate uint32,
		volsUpdate uint32,
	) error
}
