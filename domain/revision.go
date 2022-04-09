// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
//
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

package domain

import (
	"context"

	"github.com/bangumi/server/model"
)

type RevisionRepo interface {
	CountPersonRelated(ctx context.Context, personID model.PersonIDType) (int64, error)

	ListPersonRelated(
		ctx context.Context, personID model.PersonIDType, limit int, offset int,
	) ([]model.Revision, error)

	GetPersonRelated(ctx context.Context, id model.IDType) (model.Revision, error)

	CountSubjectRelated(ctx context.Context, id model.SubjectIDType) (int64, error)

	ListSubjectRelated(
		ctx context.Context, id model.SubjectIDType, limit int, offset int,
	) ([]model.Revision, error)

	GetSubjectRelated(ctx context.Context, id model.IDType) (model.Revision, error)
}
