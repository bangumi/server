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

package ctrl

import (
	"context"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger/log"
)

type UpdateCollectionRequest struct {
	VolStatus uint32
	EpStatus  uint32
	Type      model.SubjectCollection
}

func (c Ctrl) UpdateCollection(
	ctx context.Context,
	u domain.Auth,
	subjectID model.SubjectID,
	req UpdateCollectionRequest,
) error {
	c.log.Info("try to update collection", log.SubjectID(subjectID), log.UserID(u.ID), zap.Reflect("'req", req))

	err := c.collection.UpdateSubjectCollection(ctx, u.ID, subjectID, domain.SubjectCollectionUpdate{
		VolStatus: req.VolStatus,
		EpStatus:  req.EpStatus,
		Type:      req.Type,
	})
	if err != nil {
		c.log.Error("failed to update user collection info", zap.Error(err))
		return errgo.Wrap(err, "collectionRepo.UpdateSubjectCollection")
	}

	return nil
}
