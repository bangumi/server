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

type TimeLine struct {
	Related  string
	Memo     TimeLineContent
	Image    []TimelineImage
	ID       TimeLineID
	UID      uint32
	Replies  uint32
	Dateline uint32
	Cat      uint16
	Type     TimeLineType
	Batch    uint8
	Source   uint8
}

type TimeLineContent struct {
	Text string
}

type TimelineImage struct {
	Cat       *int64  `json:"cat"`
	GroupID   *string `json:"grp_id"`
	GroupName *string `json:"grp_name"`
	Name      *string `json:"name"`
	Title     *string `json:"title"`
	ID        *string `json:"id"`
	UserID    *string `json:"uid"`
	SubjectID *string `json:"subject_id"`
	Images    *string `json:"images"`
}

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
