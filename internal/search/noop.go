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
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/internal/model"
)

var _ Client = NoopClient{}

type NoopClient struct {
}

func (n NoopClient) Handle(c echo.Context) error {
	return c.String(http.StatusOK, "search is not enable")
}

func (n NoopClient) OnSubjectUpdate(ctx context.Context, id model.SubjectID) error {
	return nil
}

func (n NoopClient) OnSubjectDelete(ctx context.Context, id model.SubjectID) error {
	return nil
}

func (n NoopClient) Close() {
}
