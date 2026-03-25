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
	"github.com/labstack/echo/v5"

	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/tag"
)

type Subject struct {
	person     person.Service
	episode    episode.Repo
	personRepo person.Repo
	subject    subject.CachedRepo
	tag        tag.CachedRepo
	c          character.Repo
}

func New(
	p person.Service,
	subject subject.CachedRepo,
	personRepo person.Repo,
	c character.Repo,
	episode episode.Repo,
	tag tag.CachedRepo,
) (Subject, error) {
	return Subject{
		c:          c,
		episode:    episode,
		personRepo: personRepo,
		subject:    subject,
		person:     p,
		tag:        tag,
	}, nil
}

func (h *Subject) Routes(g *echo.Group) {
	g.GET("/subjects", h.Browse)
	g.GET("/subjects/:id", h.Get)
	g.GET("/subjects/:id/image", h.GetImage)
	g.GET("/subjects/:id/persons", h.GetRelatedPersons)
	g.GET("/subjects/:id/subjects", h.GetRelatedSubjects)
	g.GET("/subjects/:id/characters", h.GetRelatedCharacters)
}
