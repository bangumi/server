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

package cache

import (
	"github.com/bytedance/sonic"
	"github.com/trim21/errgo"
)

func marshalBytes(v any) ([]byte, error) {
	b, err := sonic.Marshal(v)
	if err != nil {
		return nil, errgo.Wrap(err, "sonic.Marshal")
	}

	return b, nil
}

func unmarshalBytes(b []byte, v any) error {
	err := sonic.Unmarshal(b, v)
	if err != nil {
		return errgo.Wrap(err, "sonic.Unmarshal")
	}

	return nil
}
