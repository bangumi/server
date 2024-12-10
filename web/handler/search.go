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
	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/internal/search"
)

func (h Handler) SearchSubjects(c echo.Context) error {
	return h.search.Handle(c, search.SearchTargetSubject) //nolint:wrapcheck
}

func (h Handler) SearchCharacters(c echo.Context) error {
	return h.search.Handle(c, search.SearchTargetCharacter) //nolint:wrapcheck
}

func (h Handler) SearchPersons(c echo.Context) error {
	return h.search.Handle(c, search.SearchTargetPerson) //nolint:wrapcheck
}
