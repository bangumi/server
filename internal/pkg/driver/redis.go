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

package driver

import (
	"fmt"
	"net/url"

	"github.com/redis/rueidis"

	"github.com/bangumi/server/config"
)

func NewRueidisClient(c config.AppConfig) (rueidis.Client, error) {
	u, err := url.Parse(c.RedisURL)
	if err != nil {
		return nil, err
	}

	password, _ := u.User.Password()
	cli, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%s", u.Hostname(), u.Port())},
		Password:    password,
		// 1<<2 = 4 connection for each node
		PipelineMultiplex: 2,
	})
	if err != nil {
		return cli, err
	}

	return cli, nil
}
