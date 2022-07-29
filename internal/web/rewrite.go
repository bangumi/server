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

package web

import (
	"regexp"
)

// rewrite openapi path params to fiber path params
// rewriteOpenapiPath("/users/-/collections/{subject_id}/episodes") => "/users/-/collections/:subject_id/episodes".
func rewriteOpenapiPath(path string) string {
	return regexp.MustCompile(`\{.*}`).ReplaceAllStringFunc(path, func(s string) string {
		return ":" + s[1:len(s)-1]
	})
}
