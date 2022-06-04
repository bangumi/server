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

package oauth

import "context"

type Manager interface {
	GetClientByID(ctx context.Context, clientIDs ...string) (map[string]Client, error)
}

type Client struct {
	ID          string
	Secret      string
	RedirectURI string
	GrantTypes  string
	Scope       string
	AppName     string
	UserID      uint32
	AppID       uint32
}
