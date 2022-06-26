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

package model

import "time"

type Group struct {
	CreatedAt   time.Time
	Name        string
	Icon        string
	Description string
	Title       string
	ID          GroupID
	NSFW        bool
	MemberCount int64
}

type GroupMember struct {
	JoinAt time.Time
	UserID UserID
	Mod    bool
}
