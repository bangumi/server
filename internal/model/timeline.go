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

type TimeLine struct {
	Related string
	Memo    TimeLineMemo
	Image   TimeLineImages

	ID   TimeLineID
	UID  UserID
	Cat  uint16 // Category
	Type TimeLineType

	Batch    uint8
	Source   uint8
	Replies  uint32
	Dateline uint32
}

type TimeLineMemo struct {
	*TimeLineDoujinMemo
	*TimeLineMonoMemo
	*TimeLineRelationMemo
	*TimeLineIndexMemo
	*TimeLineSubjectMemo
	*TimeLineWikiMemo
	*TimeLineBlogMemo
	*TimeLineGroupMemo
	*TimeLineSayMemo
	*TimeLineProgressMemo
}

type TimeLineDoujinMemo struct {
	ID    string
	Name  string
	Title string
}

type TimeLineMonoMemo struct {
	Name string `php:"name"`
	Cat  int    `php:"cat"`
	ID   int    `php:"id"`
}

type TimeLineRelationMemo struct {
	UserID   string
	Username string
	Nickname string
}

type TimeLineIndexMemo struct {
	ID          string
	Title       string
	Description string
}

type TimeLineSubjectMemo struct {
	ID             string
	TypeID         string
	Name           string
	NameCN         string
	Series         string
	CollectComment string
	CollectRate    int
}

type TimeLineWikiMemo struct {
	SubjectName   string
	SubjectNameCN string
	SubjectID     int
}

type TimeLineBlogMemo struct {
	EntryTitle       string
	EntryDescription string
	EntryID          int
}

type TimeLineGroupMemo struct {
	ID    string
	Name  string
	Title string
	Desc  string
}

type TimeLineSayMemo struct {
	*TimeLineSay
	*TimeLineSayEdit
}

type TimeLineSay string

type TimeLineSayEdit struct {
	Before string
	After  string
}

type TimeLineProgressMemo struct {
	EpName        *string
	VolsTotal     *string
	SubjectName   *string
	EpsUpdate     *int
	VolsUpdate    *int
	EpsTotal      *int
	EpSort        *int
	EpID          *EpisodeID
	SubjectID     *SubjectID
	SubjectTypeID *SubjectType
}

type TimeLineImage struct {
	Cat       *int64
	GroupID   *string
	GroupName *string
	Name      *string
	Title     *string
	ID        *string
	UserID    *string
	SubjectID *string
	Images    *string
}

type TimeLineImages []TimeLineImage
