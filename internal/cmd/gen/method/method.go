// Copyright (c) 2021-2022 Trim21 <trim21.me@gmail.com>
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

package method

import (
	"gorm.io/gen"
)

type PersonField interface {

	// GetPerson get person's extra field.
	//
	// where(prsn_id=@id AND prsn_cat="prsn")
	GetPerson(id uint32) (gen.T, error)

	// GetCharacter get character's extra field.
	//
	// where(prsn_id=@id AND prsn_cat="crt")
	GetCharacter(id uint32) (gen.T, error)
}

type Person interface {
	// Get a person from database.
	//
	// where(prsn_id=@id)
	Get(id uint32) (gen.T, error)
}

type Character interface {
	// Get a Character from database.
	//
	// where(crt_id=@id)
	Get(id uint32) (gen.T, error)
}

type SubjectRelation interface {
	// GetBySubjectID find all relation of a repository.
	//
	// where(rlt_subject_id = @id)
	GetBySubjectID(id uint32) ([]gen.T, error)
}

type PersonSubjects interface {
	// GetBySubject ...
	//
	// where(subject_id=@id)
	GetBySubject(id uint32) ([]gen.T, error)

	// GetByPerson
	//
	// where(prsn_id=@id)
	GetByPerson(id uint32) ([]gen.T, error)
}

type CharacterSubjects interface {
	// GetBySubject ...
	//
	// where(subject_id=@id)
	GetBySubject(id uint32) ([]gen.T, error)

	// GetByCharacter ...
	//
	// where(crt_id=@id)
	GetByCharacter(id uint32) ([]gen.T, error)
}

type Subject interface {
	// GetByID ...
	//
	// where(subject_id=@id)
	GetByID(id uint32) (gen.T, error)

	// GetByIDs ...
	//
	// where(subject_id IN @ids)
	GetByIDs(ids []uint32) ([]gen.T, error)
}

type SubjectField interface {
	// GetByID ...
	//
	// where(field_sid=@id)
	GetByID(id uint32) (gen.T, error)
}

type Episode interface {
	// GetByID ...
	//
	// where(ep_ep_id=@id)
	GetByID(id uint32) (gen.T, error)

	// GetBySubjectID ...
	//
	// where(ep_subject_id=@id ORDER BY `ep_disc`,`ep_type`,`ep_sort`)
	GetBySubjectID(id uint32) ([]gen.T, error)

	// GetFirst get first episode of a repository.
	// where(ep_subject_id=@subjectID ORDER BY `ep_disc`,`ep_type`,`ep_sort` LIMIT 1)
	GetFirst(subjectID uint32) (gen.T, error)

	// CountBySubjectID ...
	//
	// sql(select count(ep_id) from chii_episodes where ep_subject_id=@id)
	CountBySubjectID(id uint32) (int64, error)
}

type Member interface {
	// GetByID get a user should not be used as authorization.
	//
	// where(uid=@id)
	GetByID(id uint32) (gen.T, error)
}

type SubjectRevision interface{}
