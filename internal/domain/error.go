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

package domain

import (
	"errors"

	"github.com/bangumi/server/internal/pkg/errgo"
)

// ErrNotFound should be returned when a repo or service can't find an authorization.
var ErrNotFound = errors.New("can't find item")

var ErrSubjectNotFound = errgo.Msg(ErrNotFound, "subject not found")
var ErrPersonNotFound = errgo.Msg(ErrNotFound, "person not found")
var ErrCharacterNotFound = errgo.Msg(ErrNotFound, "character not found")
var ErrEpisodeNotFound = errgo.Msg(ErrNotFound, "episode not found")
