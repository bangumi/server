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

package metrics

import (
	"github.com/redis/go-redis/v9"
	redisprom "github.com/trim21/go-redis-prometheus"
)

func RedisHook(instance string) redis.Hook {
	return redisprom.NewHook(
		redisprom.WithNamespace("chii"),
		redisprom.WithDurationBuckets([]float64{.001, .002, .003, .004, .005, .0075, .01, .05, .1, .3, .5, .75, 1, 2}),
		redisprom.WithInstanceName(instance),
	)
}
