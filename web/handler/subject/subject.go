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
	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/subject"
)

type Subject struct {
	ctrl    ctrl.Ctrl
	person  person.Service
	subject subject.Repo
	c       character.Repo
}

func New(
	p person.Service,
	ctrl ctrl.Ctrl,
	subject subject.Repo,
	c character.Repo,
) (Subject, error) {
	return Subject{
		ctrl:    ctrl,
		c:       c,
		subject: subject,
		person:  p,
	}, nil
}
