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
	"encoding/json"

	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
)

func (e *eventHandler) OnSubject(ctx context.Context, key json.RawMessage, payload Payload) error {
	var k SubjectKey
	if err := json.Unmarshal(key, &k); err != nil {
		return err
	}

	return e.onSubjectChange(ctx, k.ID, payload.Op)
}

func (e *eventHandler) OnSubjectField(ctx context.Context, key json.RawMessage, payload Payload) error {
	var k SubjectFieldKey
	if err := json.Unmarshal(key, &k); err != nil {
		return err
	}

	return e.onSubjectChange(ctx, k.ID, payload.Op)
}

func (e *eventHandler) onSubjectChange(ctx context.Context, subjectID model.SubjectID, op string) error {
	switch op {
	case opCreate:
		if err := e.search.OnSubjectAdded(ctx, subjectID); err != nil {
			return errgo.Wrap(err, "search.OnSubjectAdded")
		}
	case opUpdate, opSnapshot:
		if err := e.search.OnSubjectUpdate(ctx, subjectID); err != nil {
			return errgo.Wrap(err, "search.OnSubjectUpdate")
		}
	case opDelete:
		if err := e.search.OnSubjectDelete(ctx, subjectID); err != nil {
			return errgo.Wrap(err, "search.OnSubjectDelete")
		}
	default:
		e.log.Warn("unexpected operator", zap.String("op", op))
	}

	return nil
}

type SubjectKey struct {
	ID model.SubjectID `json:"subject_id"`
}

type SubjectFieldKey struct {
	ID model.SubjectID `json:"field_sid"`
}
