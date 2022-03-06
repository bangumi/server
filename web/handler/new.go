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
	"go.uber.org/zap"

	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/domain"
)

func New(
	s domain.SubjectRepo,
	c domain.CharacterRepo,
	p domain.PersonRepo,
	a domain.AuthRepo,
	e domain.EpisodeRepo,
	user domain.UserRepo,
	cache cache.Generic,
	log *zap.Logger,
) Handler {
	return Handler{
		cache: cache,
		log:   log,
		p:     p,
		s:     s,
		a:     a,
		u:     user,
		e:     e,
		c:     c,
	}
}

type Handler struct {
	// replace it with service, when it's too complex. Just use a repository currently.
	s     domain.SubjectRepo
	p     domain.PersonRepo
	a     domain.AuthRepo
	e     domain.EpisodeRepo
	c     domain.CharacterRepo
	u     domain.UserRepo
	cache cache.Generic
	log   *zap.Logger
}
