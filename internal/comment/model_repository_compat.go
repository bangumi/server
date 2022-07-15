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

package comment

import (
	"time"

	"github.com/bangumi/server/internal/model"
)

func wrapCommentDao[T mysqlComment](data []T, err error) ([]mysqlComment, error) {
	if err != nil {
		return nil, err
	}

	var s = make([]mysqlComment, len(data))
	for i, item := range data {
		s[i] = item
	}

	return s, nil
}

type mysqlComment interface {
	CommentID() model.CommentID
	CreatorID() model.UserID
	IsSubComment() bool
	CreateAt() time.Time
	GetContent() string
	GetState() uint8
	RelatedTo() model.CommentID
	GetMentionedID() model.UserID
}
