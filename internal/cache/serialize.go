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
	"github.com/goccy/go-json"

	"github.com/bangumi/server/internal/pkg/errgo"
)

func marshalBytes(v any) ([]byte, error) {
	b, err := json.MarshalWithOption(v, json.DisableHTMLEscape(), json.DisableNormalizeUTF8())
	if err != nil {
		return nil, errgo.Wrap(err, "json.Marshal")
	}

	return b, nil
}

func unmarshalBytes(b []byte, v any) error {
	err := json.UnmarshalNoEscape(b, v)
	if err != nil {
		return errgo.Wrap(err, "json.Unmarshal")
	}

	return nil
}
