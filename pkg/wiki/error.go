// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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
	Is(error) bool
} = parseError{}

type parseError struct {
	err  error
	line string
	lino int
}

func (p parseError) Error() string {
	return p.err.Error() + "\nline: " + strconv.Itoa(p.lino) + " " + strconv.Quote(p.line)
}

func (p parseError) Unwrap() error {
	return p.err
}

func (p parseError) Is(err error) bool {
	return p.err == err //nolint: goerr113,errorlint
}

func wrapError(err error, lino int, line string) error {
	return parseError{
		line: line,
		lino: lino,
		err:  err,
	}
}
