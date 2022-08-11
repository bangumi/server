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

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

func main() {
	// Connect MySQL at 127.0.0.1:3306, with user root, an empty password and database test
	// conn, err := client.Connect("192.168.1.3:3306", "user", "password", "bangumi")
	// if err != nil {
	// 	panic(err)
	// }
	//
	// err = conn.Ping()
	// if err != nil {
	// 	panic(err)
	// }

	cfg := replication.BinlogSyncerConfig{
		ServerID:        3,
		Flavor:          "mysql",
		Host:            "192.168.1.3",
		Port:            3306,
		User:            "root",
		Password:        "secret",
		Charset:         "utf8mb4",
		RawModeEnabled:  true,
		ParseTime:       true,
		HeartbeatPeriod: time.Hour,
		ReadTimeout:     time.Minute,
	}
	syncer := replication.NewBinlogSyncer(cfg)

	// Start sync with specified binlog file and position
	streamer, err := syncer.StartSync(mysql.Position{Name: "mysql-bin", Pos: 1})
	if err != nil {
		panic(err)
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		ev, err := streamer.GetEvent(ctx)
		cancel()

		if err != nil {
			if err == context.DeadlineExceeded {
				// meet timeout
				continue
			}

			fmt.Println(err)
			continue
		}

		ev.Dump(os.Stdout)
	}
}
