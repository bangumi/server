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

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/model"
)

const TplListAccessToken = "list_access_token.gohtml"
const TplLogin = "login.gohtml"

type ListAccessToken struct {
	Clients map[string]string
	Title   string
	User    model.User
	Tokens  []*dao.OAuthAccessToken
}

func filters() map[string]interface{} {
	return map[string]interface{}{
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
	Execute(w io.Writer, name string, data interface{}) error
} = TemplateEngine{}
