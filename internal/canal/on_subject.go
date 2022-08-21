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

	"github.com/goccy/go-json"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/logger/log"
)

func (e *eventHandler) OnSubject(key json.RawMessage, payload payload) {
	var k SubjectKey
	if err := json.UnmarshalNoEscape(key, &k); err != nil {
		return
	}

	e.onSubjectChange(k.ID, payload.Op)
}

func (e *eventHandler) OnSubjectField(key json.RawMessage, payload payload) {
	var k SubjectFieldKey
	if err := json.UnmarshalNoEscape(key, &k); err != nil {
		return
	}

	e.onSubjectChange(k.ID, payload.Op)
}

func (e *eventHandler) onSubjectChange(subjectID model.SubjectID, op string) {
	switch op {
	case opCreate, opUpdate, opSnapshot:
		if err := e.search.OnSubjectUpdate(context.TODO(), subjectID); err != nil {
			e.log.Error("error when try to update search subject", zap.Error(err), log.SubjectID(subjectID))
		}

		return
	case opDelete:
		if err := e.search.OnSubjectUpdate(context.TODO(), subjectID); err != nil {
			e.log.Error("error when try to update search subject", zap.Error(err), log.SubjectID(subjectID))
		}
	}
}

type SubjectKey struct {
	ID model.SubjectID `json:"subject_id"`
}

type SubjectFieldKey struct {
	ID model.SubjectID `json:"field_sid"`
}
