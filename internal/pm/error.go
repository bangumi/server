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

package pm

import (
	"errors"
)

var ErrPmNotOwned = errors.New("not sent or received this private message")
var ErrPmDeleted = errors.New("private message deleted")
var ErrPmUserIrrelevant = errors.New("has user irrelevant message")
var ErrPmRelatedNotExists = errors.New("related private message not exists")
var ErrPmInvalidOperation = errors.New("invalid operation")
