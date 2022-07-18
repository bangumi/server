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

package auth

const (
	// UserGroupAdmin Deprecated.
	UserGroupAdmin uint8 = iota + 1
	// UserGroupBangumiAdmin Deprecated.
	UserGroupBangumiAdmin
	// UserGroupWindowAdmin Deprecated.
	UserGroupWindowAdmin
	// UserGroupQuite Deprecated.
	UserGroupQuite
	// UserGroupBanned Deprecated.
	UserGroupBanned
	_
	_
	// UserGroupCharacterAdmin Deprecated.
	UserGroupCharacterAdmin
	// UserGroupWikiAdmin Deprecated.
	UserGroupWikiAdmin
	// UserGroupNormal Deprecated.
	UserGroupNormal
	// UserGroupWikiEditor Deprecated.
	UserGroupWikiEditor
)
