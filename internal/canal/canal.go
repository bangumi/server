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

package canal

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/pkg/logger"
)

type savedPosition struct {
	Pos  mysql.Position
	Time int64
}

func NewEventHandler(appConfig config.AppConfig, r cache.Cache) (*MyEventHandler, error) {
	cfg := canal.NewDefaultConfig()
	cfg.Dump.ExecutionPath = "" // disable dump
	cfg.ServerID = appConfig.MySQLBinlogServerID
	cfg.Addr = fmt.Sprintf("%s:%s", appConfig.MySQLHost, appConfig.MySQLPort)
	cfg.User = appConfig.MySQLUserName
	cfg.Password = appConfig.MySQLPassword

	c, err := canal.NewCanal(cfg)
	if err != nil {
		return nil, err
	}

	h := &MyEventHandler{
		c:             c,
		r:             r,
		subjectUpdate: make(chan model.SubjectID),
		subjectDelete: make(chan model.SubjectID),
	}

	c.SetEventHandler(h)

	return h, nil
}

type MyEventHandler struct {
	canal.DummyEventHandler

	c *canal.Canal

	subjectUpdate chan model.SubjectID
	subjectDelete chan model.SubjectID

	r cache.Cache

	saved mysql.Position

	pos atomic.Value // savedPosition
}

func (h *MyEventHandler) OnRotate(e *replication.RotateEvent) error {
	h.pos.Store(savedPosition{
		Pos: mysql.Position{
			Name: string(e.NextLogName),
			Pos:  uint32(e.Position),
		},
		Time: time.Now().Unix(),
	})

	return nil
}

func (h *MyEventHandler) SavePosToRedis() {
	// save to redis
	for {
		time.Sleep(5 * time.Second)
		i := h.pos.Load()
		if i != nil {
			pos := i.(savedPosition)
			if pos.Pos != h.saved {
				if err := h.r.Set(context.TODO(), redisSaveKey, pos, gtime.OneDay); err != nil {
					logger.Error("failed to save canal binlog position", zap.Error(err))
					continue
				}
				fmt.Println("save binlog Pos to redis")
				h.saved = pos.Pos
			}
		}
	}
}

func (h *MyEventHandler) String() string {
	return "MyEventHandler"
}

func (h *MyEventHandler) Run() error {
	go h.OnSubjectUpdate()
	go h.OnSubjectDelete()
	go h.SavePosToRedis()
	return h.c.Run()
}

const redisSaveKey = "canal-mysql-binlog-Pos"

func Main() error {
	h, err := New()
	if err != nil {
		return err
	}

	// Start canal, blocking
	return h.Run()
}

func New() (*MyEventHandler, error) {
	var h *MyEventHandler

	err := fx.New(
		logger.FxLogger(),
		config.Module,

		// driver and connector
		fx.Provide(
			driver.NewRedisClient, logger.Copy, cache.NewRedisCache,

			NewEventHandler,
		),

		fx.Populate(&h),
	).Err()

	if err != nil {
		return nil, errgo.Wrap(err, "fx")
	}

	return h, nil
}
