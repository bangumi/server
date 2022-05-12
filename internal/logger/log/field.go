// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
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

package log

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/model"
)

func UserID(id model.UIDType) zap.Field {
	return zap.Uint32("user_id", id)
}

func GroupID(id domain.GroupID) zap.Field {
	return zap.Uint8("user_id", id)
}

func SubjectID(id model.SubjectIDType) zap.Field {
	return zap.Uint32("subject_id", id)
}
