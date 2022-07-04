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

package wiki

import "strconv"

var _ interface {
	Error() string
	Unwrap() error
} = (*SyntaxError)(nil)

type SyntaxError struct {
	Err  error
	Line string
	Lino int
}

func (p *SyntaxError) Error() string {
	return p.Err.Error() + " line: " + strconv.Itoa(p.Lino) + " " + strconv.Quote(p.Line)
}

func (p *SyntaxError) Unwrap() error {
	return p.Err
}

func wrapError(err error, lino int, line string) error {
	return &SyntaxError{
		Line: line,
		Lino: lino,
		Err:  err,
	}
}
