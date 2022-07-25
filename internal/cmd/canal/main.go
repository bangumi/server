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

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

func main() {
	// Connect MySQL at 127.0.0.1:3306, with user root, an empty password and database test
	conn, err := client.Connect("192.168.1.3:3306", "user", "password", "bangumi")
	if err != nil {
		panic(err)
	}

	err = conn.Ping()
	if err != nil {
		panic(err)
	}

	cfg := replication.BinlogSyncerConfig{
		ServerID:        100,
		Flavor:          "mysql",
		Host:            "192.168.1.3",
		Port:            3306,
		User:            "user",
		Password:        "password",
		Charset:         "utf8mb4",
		RawModeEnabled:  false,
		ParseTime:       true,
		HeartbeatPeriod: time.Hour,
		ReadTimeout:     time.Minute,
	}
	syncer := replication.NewBinlogSyncer(cfg)

	// Start sync with specified binlog file and position
	streamer, err := syncer.StartSync(mysql.Position{"search", 0})
	if err != nil {
		panic(err)
	}

	// or you can start a gtid replication like
	// streamer, _ := syncer.StartSyncGTID(gtidSet)
	// the mysql GTID set likes this "de278ad0-2106-11e4-9f8e-6edd0ca20947:1-2"
	// the mariadb GTID set likes this "0-1-100"

	// or we can use a timeout context
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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
