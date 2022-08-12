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
	"reflect"
	"sync/atomic"
	"time"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/gookit/event"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/pool"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/web/session"
)

type savedPosition struct {
	Pos  mysql.Position
	Time int64
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

func New() (*BinlogHandler, error) {
	var h *BinlogHandler

	err := fx.New(
		logger.FxLogger(),
		config.Module,

		dal.Module,

		// driver and connector
		fx.Provide(
			driver.NewMysqlConnectionPool, metrics.NewScope,
			driver.NewRedisClient, logger.Copy, cache.NewRedisCache,

			session.NewMysqlRepo, session.New,

			eventManager, NewEventHandler,
		),

		fx.Populate(&h),
	).Err()

	if err != nil {
		return nil, errgo.Wrap(err, "fx")
	}

	return h, nil
}

func NewEventHandler(
	appConfig config.AppConfig,
	log *zap.Logger,
	e *event.Manager,
	r cache.Cache,
) (*BinlogHandler, error) {
	cfg := canal.NewDefaultConfig()
	cfg.Dump.ExecutionPath = "" // disable dump
	cfg.ServerID = appConfig.MySQLBinlogServerID
	cfg.Addr = fmt.Sprintf("%s:%s", appConfig.MySQLHost, appConfig.MySQLPort)
	cfg.User = appConfig.MySQLUserName
	cfg.Password = appConfig.MySQLPassword

	c, err := canal.NewCanal(cfg)
	if err != nil {
		return nil, errgo.Wrap(err, "canal.NewCanal")
	}

	h := &BinlogHandler{
		e:   e,
		c:   c,
		log: log.Named("BinlogHandler"),
		r:   r,
	}

	c.SetEventHandler(h)

	return h, nil
}

type BinlogHandler struct {
	canal.DummyEventHandler
	r     cache.Cache
	pos   atomic.Value
	log   *zap.Logger
	c     *canal.Canal
	e     *event.Manager
	saved mysql.Position
}

func (h *BinlogHandler) OnRotate(e *replication.RotateEvent) error {
	h.pos.Store(savedPosition{
		Pos: mysql.Position{
			Name: string(e.NextLogName),
			Pos:  uint32(e.Position),
		},
		Time: time.Now().Unix(),
	})

	return nil
}

func (h *BinlogHandler) SavePosToRedis() {
	// save to redis
	for {
		time.Sleep(5 * time.Second)
		i := h.pos.Load()
		if i != nil {
			pos := i.(savedPosition) //nolint:forcetypeassert
			if pos.Pos != h.saved {
				if err := h.r.Set(context.TODO(), redisSaveKey, pos, gtime.OneDay); err != nil {
					logger.Error("failed to save canal binlog position", zap.Error(err))
					continue
				}
				h.log.Info("save binlog Pos to redis")
				h.saved = pos.Pos
			}
		}
	}
}

func (h *BinlogHandler) String() string {
	return "BinlogHandler"
}

func (h *BinlogHandler) Run() error {
	go h.SavePosToRedis()
	return errgo.Wrap(h.c.Run(), "canal.Run")
}

func (h *BinlogHandler) OnRow(e *canal.RowsEvent) error {
	if e.Table.Schema != "bangumi" {
		return nil
	}

	switch e.Header.EventType {
	case replication.WRITE_ROWS_EVENTv2,
		replication.UPDATE_ROWS_EVENTv2,
		replication.DELETE_ROWS_EVENTv2:
	default:
		return nil
	}

	switch e.Table.Name {
	case "chii_members":
		err := h.userRow(e)
		if err != nil {
			return err
		}
	case "chii_subjects":
		err := h.onSubjectRow(e)
		if err != nil {
			return err
		}
	case "chii_subject_fields":
		err := h.onSubjectFieldRow(e)
		if err != nil {
			return err
		}
	default:
		return nil
	}

	h.updatePos()

	return nil
}

func (h *BinlogHandler) updatePos() {
	h.pos.Store(savedPosition{
		Pos:  h.c.SyncedPosition(),
		Time: time.Now().Unix(),
	})
}

func (h *BinlogHandler) onSubjectRow(e *canal.RowsEvent) error {
	const idColumn = "subject_id"
	return h.subjectEventHandler(e, idColumn)
}

func (h *BinlogHandler) onSubjectFieldRow(e *canal.RowsEvent) error {
	const idColumn = "field_sid"
	return h.subjectEventHandler(e, idColumn)
}

var subjectIDPool = pool.New(func() []model.SubjectID {
	return make([]model.SubjectID, 0, 3)
})

func (h *BinlogHandler) subjectEventHandler(e *canal.RowsEvent, idColumn string) error {
	ids := subjectIDPool.Get()
	defer func() {
		subjectIDPool.Put(ids[:0])
	}()

	for i, column := range e.Table.Columns {
		if column.Name == idColumn {
			for _, row := range e.Rows {
				c := row[i]
				sid := model.SubjectID(reflect.ValueOf(c).Uint())
				ids = append(ids, sid)
			}
			break
		}
	}

	switch e.Action {
	case canal.InsertAction:
		for _, id := range ids {
			h.e.AsyncFire(NewSubjectEvent(EventSubjectCreate, id))
		}
		return nil
	case canal.DeleteAction:
		for _, id := range ids {
			h.e.AsyncFire(NewSubjectEvent(EventSubjectDelete, id))
		}
		return nil
	case canal.UpdateAction:
		for i := 1; i < len(ids); i += 2 {
			before, after := ids[i-1], ids[i]
			if before != after {
				h.e.AsyncFire(NewSubjectEvent(EventSubjectCreate, after))
				h.e.AsyncFire(NewSubjectEvent(EventSubjectDelete, before))
			} else {
				h.e.AsyncFire(NewSubjectEvent(EventSubjectUpdate, before))
			}
		}
	}

	return nil
}

func (h *BinlogHandler) userRow(e *canal.RowsEvent) error {
	switch e.Action {
	case canal.UpdateAction:

	default:
		h.log.Debug("ignore user change on insert and delete", zap.String("action", e.Action))
		return nil
	}

	idColumn := e.Table.FindColumn("uid")
	if idColumn < 0 {
		h.log.Error("unexpected column index", zap.Int("index", idColumn))
		return nil
	}

	column := e.Table.FindColumn("password_crypt")
	if column < 0 {
		h.log.Error("unexpected column index", zap.Int("index", column))
		return nil
	}

	for i := 1; i < len(e.Rows); i += 2 {
		afterPass := e.Rows[i][column]
		beforePass := e.Rows[i-1][column]
		if beforePass != afterPass {
			h.e.AsyncFire(NewUserChangePassword(model.UserID(reflect.ValueOf(e.Rows[i][idColumn]).Uint())))
		}
	}

	return nil
}
