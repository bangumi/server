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

package collection

import (
	"github.com/bangumi/server/internal/model"
)

func (s *Subject) User() model.UserID {
	return s.user
}

func (s *Subject) Rate() uint8 {
	return s.rate
}

func (s *Subject) TypeID() SubjectCollection {
	return s.typeID
}

func (s *Subject) Comment() string {
	return s.comment
}

func (s *Subject) Privacy() CollectPrivacy {
	return s.privacy
}

func (s *Subject) Tags() []string {
	return s.tags
}

func (s *Subject) Vols() uint32 {
	return s.vols
}

func (s *Subject) Eps() uint32 {
	return s.eps
}

func (s *Subject) Subject() model.SubjectID {
	return s.subject
}
