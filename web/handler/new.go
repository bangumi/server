// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
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
	s domain.SubjectService,
	c domain.CharacterService,
	p domain.PersonService,
	a domain.AuthRepo,
	e domain.EpisodeRepo,
	r domain.RevisionRepo,
	index domain.IndexRepo,
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
		i:     index,
		r:     r,
	}
}

type Handler struct {
	// replace it with service, when it's too complex. Just use a repository currently.
	s     domain.SubjectService
	p     domain.PersonService
	a     domain.AuthRepo
	e     domain.EpisodeRepo
	c     domain.CharacterService
	u     domain.UserRepo
	i     domain.IndexRepo
	r     domain.RevisionRepo
	cache cache.Generic
	log   *zap.Logger
}
