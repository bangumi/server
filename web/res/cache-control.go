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

package res

import (
	"bytes"
	"fmt"
	"time"
)

type CacheControlParams struct {
	Public bool
	MaxAge time.Duration
}

func (c CacheControlParams) String() string {
	buf := bytes.NewBuffer(make([]byte, 0, 20))

	_, _ = buf.WriteString("public, ")

	_, _ = fmt.Fprintf(buf, "max-age=%d", c.MaxAge/time.Second)

	return buf.String()
}
