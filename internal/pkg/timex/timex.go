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

package timex

import "time"

const (
	OneMinSec  = 60
	OneHourSec = 3600
	OneDaySec  = 86400
	OneWeekSec = 7 * 86400

	OneDay  = 24 * time.Hour
	OneWeek = 7 * 24 * time.Hour
)

func NowU32() uint32 {
	return uint32(time.Now().Unix())
}
