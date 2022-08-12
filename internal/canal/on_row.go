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
	"reflect"
	"time"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/replication"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/pool"
)

func (h *EventHandler) OnRow(e *canal.RowsEvent) error {
	switch e.Header.EventType {
	case replication.WRITE_ROWS_EVENTv2,
		replication.UPDATE_ROWS_EVENTv2,
		replication.DELETE_ROWS_EVENTv2:
	default:
		return nil
	}

	switch e.Table.Name {
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

func (h *EventHandler) updatePos() {
	h.pos.Store(savedPosition{
		Pos:  h.c.SyncedPosition(),
		Time: time.Now().Unix(),
	})
}

func (h *EventHandler) onSubjectRow(e *canal.RowsEvent) error {
	const idColumn = "subject_id"
	return h.subjectEventHandler(e, idColumn)
}

func (h *EventHandler) onSubjectFieldRow(e *canal.RowsEvent) error {
	const idColumn = "field_sid"
	return h.subjectEventHandler(e, idColumn)
}

var idPool = pool.New(func() []model.SubjectID {
	return make([]model.SubjectID, 0, 3)
})

func (h *EventHandler) subjectEventHandler(e *canal.RowsEvent, idColumn string) error {
	ids := idPool.Get()
	defer func() {
		idPool.Put(ids[:0])
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

	switch len(ids) {
	case 1:
		h.subjectUpdate <- ids[0]
	case 2:
		if ids[0] == ids[1] {
			h.subjectUpdate <- ids[0]
		} else {
			h.subjectDelete <- ids[0]
			h.subjectUpdate <- ids[1]
		}
	default:
		h.log.Warn("unexpected id length", zap.Int("len", len(ids)))
	}

	return nil
}
