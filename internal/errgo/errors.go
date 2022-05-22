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

package errgo

// Wrap add context to error message.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &wrapError{msg: msg, err: err}
}

func Msg(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &msgError{msg: msg, err: err}
}

type wrapError struct {
	err error
	msg string
}

func (e *wrapError) Error() string {
	return e.msg + ": " + e.err.Error()
}

func (e *wrapError) Unwrap() error {
	return e.err
}

type msgError struct {
	err error
	msg string
}

func (e *msgError) Error() string {
	return e.msg
}

func (e *msgError) Unwrap() error {
	return e.err
}
