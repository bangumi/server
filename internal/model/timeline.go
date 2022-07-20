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

package model

import (
	"fmt"

	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/trim21/go-phpserialize"
)

const (
	TimeLineTypeAll TimeLineType = iota
	TimeLineTypeSubject
	TimeLineTypeProgress
	TimeLineTypeRelation
	TimeLineTypeGroup
	TimeLineTypeSay
	TimeLineTypeWiki
	TimeLineTypeBlog
	TimeLineTypeIndex
	TimeLineTypeMono
	TimeLineTypeDoujin
	TimeLineTypeReplies
)

var ErrEmptyMemo = fmt.Errorf("empty")
var ErrNilMemo = fmt.Errorf("nil")

type TimeLine struct {
	ID   TimeLineID
	UID  UserID
	Cat  uint16 // Category
	Type TimeLineType

	Memo     TimeLineContent
	Image    TimeLineImages
	Related  string
	Replies  uint32
	Dateline uint32
	Batch    uint8
	Source   uint8
}

type TimeLineContent struct {
	*TimeLineSayMemo
	*TimeLineProgressMemo
}

type TimeLineSayMemo string

func (m *TimeLineSayMemo) FromBytes(bytes []byte) error {
	if m == nil {
		return ErrNilMemo
	}
	*m = TimeLineSayMemo(bytes)
	return nil
}

func (m *TimeLineSayMemo) Bytes() ([]byte, error) {
	if m == nil {
		return nil, ErrEmptyMemo
	}
	return []byte(*m), nil
}

type TimeLineProgressMemo struct {
	EpName        *string      `json:"ep_name,omitempty" php:"ep_name,omitempty"`
	VolsTotal     *string      `json:"vols_total,omitempty" php:"vols_total,omitempty"`
	SubjectName   *string      `json:"subject_name,omitempty" php:"subject_name,omitempty"`
	EpsUpdate     *int         `json:"eps_update,omitempty" php:"eps_update,omitempty,string"`
	VolsUpdate    *int         `json:"vols_update,omitempty" php:"vols_update,omitempty,string"`
	EpsTotal      *int         `json:"eps_total,omitempty" php:"eps_total,omitempty,string"`
	EpSort        *int         `json:"ep_sort,omitempty" php:"ep_sort,omitempty,string"`
	EpID          *EpisodeID   `json:"ep_id,omitempty" php:"ep_id,omitempty,string"`
	SubjectID     *SubjectID   `json:"subject_id,omitempty" php:"subject_id,omitempty,string"`
	SubjectTypeID *SubjectType `json:"subject_type_id,omitempty" php:"subject_type_id,omitempty,string"`
}

func (m *TimeLineProgressMemo) Bytes() ([]byte, error) {
	if m == nil {
		return nil, ErrEmptyMemo
	}
	result, err := phpserialize.Marshal(m)
	return result, errgo.Wrap(err, "phpserialize.Marshal")
}

func (m *TimeLineProgressMemo) FromBytes(b []byte) error {
	if m == nil {
		return ErrNilMemo
	}
	if err := phpserialize.Unmarshal(b, m); err != nil {
		return errgo.Wrap(err, "phpserialize.Unmarshal")
	}
	return nil
}

type TimeLineImage struct {
	Cat       *int64  `json:"cat,omitempty" php:"cat,omitempty"`
	GroupID   *string `json:"grp_id,omitempty" php:"grp_id,omitempty"`
	GroupName *string `json:"grp_name,omitempty" php:"grp_name,omitempty"`
	Name      *string `json:"name,omitempty" php:"name,omitempty"`
	Title     *string `json:"title,omitempty" php:"title,omitempty"`
	ID        *string `json:"id,omitempty" php:"id,omitempty"`
	UserID    *string `json:"uid,omitempty" php:"uid,omitempty"`
	SubjectID *string `json:"subject_id,omitempty" php:"subject_id,omitempty"`
	Images    *string `json:"images,omitempty" php:"images,omitempty"`
}

func (i *TimeLineImage) Bytes() ([]byte, error) {
	if i == nil {
		return nil, ErrEmptyMemo
	}
	result, err := phpserialize.Marshal(i)
	return result, errgo.Wrap(err, "phpserialize.Marshal")
}

type TimeLineImages []TimeLineImage

func (is TimeLineImages) Bytes() ([]byte, error) {
	if len(is) == 0 {
		return nil, ErrEmptyMemo
	}
	if len(is) == 1 {
		return is[0].Bytes()
	}

	result, err := phpserialize.Marshal(is)
	return result, errgo.Wrap(err, "phpserialize.Marshal")
}
