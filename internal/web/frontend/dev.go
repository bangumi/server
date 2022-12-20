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

//nolint:wrapcheck
package frontend

// dev file to avoid rebuild whole application when only editing template and static files.

import (
	"html/template"
	"io"

	"github.com/Masterminds/sprig/v3"
)

type devEngine struct {
}

func newDevTemplateEngine() (TemplateEngine, error) {
	return devEngine{}, nil
}

func (e devEngine) Execute(w io.Writer, name string, data any) error {
	t, err := template.New("").Funcs(filters()).Funcs(sprig.FuncMap()).
		ParseGlob("./internal/web/frontend/templates/**.gohtml")
	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, name, data)
}
