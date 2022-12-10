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

package req

type PrivateMessageMarkRead struct {
	ID uint32 `json:"id" validate:"required,gt=0"`
}

type PrivateMessageCreate struct {
	Title       string   `json:"title" validate:"required,gt=0,lte=100"`
	Content     string   `json:"content" validate:"required,gt=0,lte=1000"`
	RelatedID   *uint32  `json:"related_id" validate:"omitempty,gt=0"`
	ReceiverIDs []uint32 `json:"receiver_ids" validate:"required,gt=0,dive,gt=0"`
}

type PrivateMessageDelete struct {
	IDs []uint32 `json:"ids" validate:"required,gt=0,dive,gt=0"`
}
