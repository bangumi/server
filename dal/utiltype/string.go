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

package utiltype

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"html"
	"reflect"

	"github.com/samber/lo"
)

var _ driver.Valuer = HTMLEscapedString("")
var _ sql.Scanner = lo.ToPtr(HTMLEscapedString(""))

type HTMLEscapedString string

func (s *HTMLEscapedString) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*s = HTMLEscapedString(html.UnescapeString(v))
		return nil
	case []byte:
		*s = HTMLEscapedString(html.UnescapeString(string(v)))
		return nil
	case sql.RawBytes:
		*s = HTMLEscapedString(html.UnescapeString(string(v)))
		return nil
	}

	//nolint:goerr113
	return fmt.Errorf("utiltype.HTMLEscapedString: unsupported input type %s", reflect.TypeOf(src).String())
}

func (s HTMLEscapedString) Value() (driver.Value, error) {
	return html.EscapeString(string(s)), nil
}
