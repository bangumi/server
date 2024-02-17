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
	"strings"

	"github.com/samber/lo"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/dam"
	"github.com/bangumi/server/internal/pkg/generic/slice"
)

func NewSubjectCollection(
	subject model.SubjectID,
	user model.UserID,
	rate uint8,
	typeID SubjectCollection,
	comment string,
	privacy CollectPrivacy,
	tags []string,
	vols uint32,
	eps uint32,
) (*Subject, error) {
	if subject == 0 {
		return nil, errgo.Wrap(gerr.ErrInvalidData, "empty subject id")
	}
	if user == 0 {
		return nil, errgo.Wrap(gerr.ErrInvalidData, "empty user id")
	}

	if rate > 10 {
		return nil, errgo.Wrap(gerr.ErrInvalidData, "rate overflow")
	}

	switch privacy {
	case CollectPrivacyNone, CollectPrivacySelf, CollectPrivacyBan:
	default:
		return nil, errgo.Wrap(gerr.ErrInvalidData, "rate overflow")
	}

	return &Subject{
		subject: subject,
		user:    user,
		rate:    rate,
		typeID:  typeID,
		comment: comment,
		privacy: privacy,
		tags:    tags,
		vols:    vols,
		eps:     eps,
	}, nil
}

func NewEmptySubjectCollection(subject model.SubjectID, user model.UserID) *Subject {
	return &Subject{
		subject: subject,
		user:    user,
	}
}

type Subject struct {
	subject model.SubjectID
	user    model.UserID
	rate    uint8
	typeID  SubjectCollection
	comment string
	privacy CollectPrivacy
	tags    []string
	vols    uint32
	eps     uint32
}

func (s *Subject) ShadowBan(v bool) {
	if v {
		s.privacy = CollectPrivacyBan
	} else if s.privacy == CollectPrivacyBan {
		s.privacy = CollectPrivacySelf
	}
}

func (s *Subject) MakePrivate() {
	if s.privacy != CollectPrivacyBan {
		s.privacy = CollectPrivacySelf
	}
}

func (s *Subject) MakePublic() {
	if s.privacy != CollectPrivacyBan {
		s.privacy = CollectPrivacyNone
	}
}

func (s *Subject) UpdateComment(comment string) error {
	comment = strings.TrimSpace(comment)

	if comment == "" {
		s.comment = ""
		return nil
	}

	if !dam.AllPrintableChar(comment) {
		return gerr.ErrInvisibleChar
	}

	s.comment = comment

	return nil
}

func (s *Subject) UpdateTags(tags []string) error {
	if tags == nil {
		s.tags = nil
		return nil
	}

	tags = slice.Map(tags, strings.TrimSpace)

	if lo.ContainsBy(tags, func(item string) bool { return !dam.AllPrintableChar(item) }) {
		return gerr.ErrInvisibleChar
	}

	s.tags = lo.Uniq(tags)

	return nil
}

func (s *Subject) UpdateType(r SubjectCollection) {
	s.typeID = r
}

func (s *Subject) UpdateRate(r uint8, state SubjectCollection) error {
	if r > 10 {
		return errgo.Wrap(gerr.ErrInput, "rate overflow")
	}

	if state == SubjectCollectionWish {
		s.rate = 0
	} else {
		s.rate = r
	}

	return nil
}

func (s *Subject) UpdateVols(v uint32) {
	s.vols = v
}

func (s *Subject) UpdateEps(v uint32) {
	s.eps = v
}
