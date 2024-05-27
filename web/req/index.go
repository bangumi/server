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

import "github.com/bangumi/server/internal/model"

type IndexBasicInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type IndexAddSubject struct {
	SubjectID model.SubjectID `json:"subject_id"`
	*IndexSubjectInfo
}

type IndexSubjectInfo struct {
	SortKey uint32 `json:"sort"`
	Comment string `json:"comment"`
}

type IndexComment struct {
	ID      model.CommentID `json:"id"`
	FieldID model.IndexID   `json:"field_id"`
	Comment string          `json:"comment"`
}
