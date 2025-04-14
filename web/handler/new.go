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

package handler

import (
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/revision"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/user"
)

func New(
	r revision.Repo,
	subject subject.CachedRepo,
	search search.Handler,
	u user.Repo,
	episode episode.Repo,
) Handler {
	return Handler{
		episode: episode,
		u:       u,
		subject: subject,
		search:  search,
		r:       r,
	}
}

type Handler struct {
	episode episode.Repo
	r       revision.Repo
	subject subject.CachedRepo
	u       user.Repo
	search  search.Handler
}
