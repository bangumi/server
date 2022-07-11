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

package frontend

import (
	"html/template"
	"io"
	"math/rand"
	"time"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/oauth"
)

const TplLogin = "login.gohtml"
const TplListAccessToken = "list-access-token.gohtml"
const TplCreateAccessToken = "create-access-token.gohtml"

func filters() map[string]any {
	return map[string]any{
		"RandInt": func() int {
			return rand.Intn(6) //nolint:gomnd,gosec
		},
		"FormatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	}
}

type TemplateEngine struct {
	t *template.Template
}

var _ interface {
	Execute(w io.Writer, name string, data any) error
} = TemplateEngine{}

type Login struct {
	Title string
	User  model.User
}

type ListAccessToken struct {
	Clients map[string]oauth.Client
	Title   string
	User    model.User
	Tokens  []domain.AccessToken
}

type CreateAccessToken struct {
	User model.User
}
