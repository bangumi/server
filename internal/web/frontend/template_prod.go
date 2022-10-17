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

//go:build !dev

package frontend

import (
	"embed"
	"html/template"
	"io"

	"github.com/Masterminds/sprig/v3"

	"github.com/bangumi/server/internal/pkg/errgo"
)

//go:embed templates
var templateFS embed.FS

func NewTemplateEngine() (TemplateEngine, error) {
	t, err := template.New("").Funcs(filters()).Funcs(sprig.FuncMap()).ParseFS(templateFS, "templates/**.gohtml")
	if err != nil {
		return TemplateEngine{}, errgo.Wrap(err, "template")
	}

	return TemplateEngine{t: t}, nil
}

func (e TemplateEngine) Execute(w io.Writer, name string, data any) error {
	return e.t.ExecuteTemplate(w, name, data) //nolint:wrapcheck
}
