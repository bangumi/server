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

package search

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/accessor"
)

var _ Client = NopeClient{}

type NopeClient struct {
}

func (n NopeClient) Handle(ctx *fiber.Ctx, auth *accessor.Accessor) error {
	return ctx.SendString("search is not enable")
}

func (n NopeClient) OnSubjectUpdate(ctx context.Context, id model.SubjectID) error {
	return nil
}

func (n NopeClient) OnSubjectDelete(ctx context.Context, id model.SubjectID) error {
	return nil
}
