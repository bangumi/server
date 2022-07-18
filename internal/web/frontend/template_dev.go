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

//go:build dev

package frontend

// dev file to avoid rebuild whole application when only editing template and static files.

import (
	"html/template"
	"io"
	"os"

	"github.com/Masterminds/sprig/v3"
)

var StaticFS = os.DirFS("./internal/web/frontend") //nolint:gochecknoglobals

func NewTemplateEngine() (TemplateEngine, error) {
	return TemplateEngine{}, nil
}

func (e TemplateEngine) Execute(w io.Writer, name string, data any) error {
	t, err := template.New("").Funcs(filters()).Funcs(sprig.FuncMap()).
		ParseGlob("./internal/web/frontend/templates/**.gohtml")
	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, name, data) //nolint:wrapcheck
}
