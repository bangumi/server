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

package graph

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/subject"
)

var userKeyStub = "user_key"
var CurrentUserKey = &userKeyStub

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	log     *zap.Logger
	subject subject.Repo
	episode domain.EpisodeRepo
}

func NewResolver(
	episode domain.EpisodeRepo,
	subjectRepo subject.Repo,
	log *zap.Logger,
) Resolver {
	return Resolver{
		episode: episode,
		subject: subjectRepo,
		log:     log.Named("GraphQL.resolver"),
	}
}
