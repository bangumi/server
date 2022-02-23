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

package handler

import (
	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/domain"
)

func New(
	s domain.SubjectRepo,
	a domain.AuthRepo,
	e domain.EpisodeRepo,
	cache cache.Generic,
) Handler {
	return Handler{
		cache: cache,
		s:     s,
		a:     a,
		e:     e,
	}
}

type Handler struct {
	// replace it with service, when it's too complex. Just use a repository currently.
	s     domain.SubjectRepo
	a     domain.AuthRepo
	e     domain.EpisodeRepo
	cache cache.Generic
}
