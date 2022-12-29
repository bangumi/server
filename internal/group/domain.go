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

package group

import (
	"context"

	"github.com/bangumi/server/internal/model"
)

type Repo interface {
	GetByName(ctx context.Context, name string) (model.Group, error)
	GetByID(ctx context.Context, id model.GroupID) (model.Group, error)

	CountMembersByID(ctx context.Context, id model.GroupID, memberType MemberType) (int64, error)
	ListMembersByID(
		ctx context.Context, id model.GroupID, memberType MemberType, limit, offset int,
	) ([]model.GroupMember, error)
}

type MemberType uint8

const (
	MemberAll MemberType = 1 << iota / 2
	MemberMod
	MemberNormal
)
