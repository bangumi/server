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
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/replication"

	"github.com/bangumi/server/internal/model"
)

func (h *MyEventHandler) OnRow(e *canal.RowsEvent) error {
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

	fmt.Println(e.Header.LogPos)
	h.updatePos()

	return nil
}

func (h *MyEventHandler) updatePos() {
	h.pos.Store(savedPosition{
		Pos:  h.c.SyncedPosition(),
		Time: time.Now().Unix(),
	})
}

func (h *MyEventHandler) onSubjectRow(e *canal.RowsEvent) error {
	for _, row := range e.Rows {
		for i, c := range row {
			if e.Table.Columns[i].Name == "id" {
				sid := model.SubjectID(reflect.ValueOf(c).Uint())
				fmt.Println("on subject row", sid)
				switch e.Action {
				case canal.DeleteAction:
					h.subjectDelete <- sid
				case canal.InsertAction, canal.UpdateAction:
					h.subjectUpdate <- sid
				default:
					panic("unexpected action " + e.Action)
				}
				break
			}
		}
	}

	return nil
}

func (h *MyEventHandler) onSubjectFieldRow(e *canal.RowsEvent) error {
	fmt.Println(e.Action)
	fmt.Println(e.Table)
	fmt.Println(e.Header)
	e.Header.Dump(os.Stdout)
	fmt.Println(len(e.Rows))

	switch e.Header.EventType {
	case replication.DELETE_ROWS_EVENTv0:
	case replication.DELETE_ROWS_EVENTv1:
	case replication.DELETE_ROWS_EVENTv2:

	case replication.UPDATE_ROWS_EVENTv0:
	case replication.UPDATE_ROWS_EVENTv1:
	case replication.UPDATE_ROWS_EVENTv2:

	case replication.WRITE_ROWS_EVENTv0:
	case replication.WRITE_ROWS_EVENTv1:
	case replication.WRITE_ROWS_EVENTv2:
	}

	for _, row := range e.Rows {
		for i, c := range row {
			if e.Table.Columns[i].Name == "field_sid" {
				sid := model.SubjectID(reflect.ValueOf(c).Uint())
				fmt.Println("on subject field row", sid)
				switch e.Action {
				case canal.DeleteAction:
					h.subjectDelete <- sid
				case canal.InsertAction, canal.UpdateAction:
					h.subjectUpdate <- sid
				default:
					panic("unexpected action " + e.Action)
				}
				break
			}
		}
	}

	return nil
}
