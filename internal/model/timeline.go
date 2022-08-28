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
	ID  TimeLineID
	UID UserID

	Related  string
	Batch    uint8
	Source   uint8
	Replies  uint32
	Dateline uint32
	Image    TimeLineImages

	Cat  uint16 // Category
	Type TimeLineType
	Memo TimeLineMemo
}

type TimeLineMemo struct {
	*TimeLineRelationMemo
	*TimeLineGroupMemo
	*TimeLineWikiMemo
	*TimeLineSubjectMemo
	*TimeLineProgressMemo
	*TimeLineSayMemo
	*TimeLineBlogMemo
	*TimeLineIndexMemo
	*TimeLineMonoMemo
	*TimeLineDoujinMemo
}

//nolint:gomnd,gocyclo
func (tl *TimeLine) FillCatAndType() *TimeLine {
	m := tl.Memo
	if m.TimeLineRelationMemo != nil {
		return setCatAndType(tl, 1, 2)
	}
	if m.TimeLineGroupMemo != nil {
		return setCatAndType(tl, 1, 3)
	}
	if m.TimeLineWikiMemo != nil {
		return setCatAndType(tl, 2, 0)
	}
	if m.TimeLineSubjectMemo != nil {
		return setCatAndType(tl, 3, 0)
	}
	if m.TimeLineProgressMemo != nil {
		return setCatAndType(tl, 4, 0)
	}
	if m.TimeLineSayMemo != nil {
		if m.TimeLineSayMemo.TimeLineSayEdit != nil {
			return setCatAndType(tl, 5, 2)
		}
		return setCatAndType(tl, 5, 0)
	}
	if m.TimeLineBlogMemo != nil {
		return setCatAndType(tl, 6, 0)
	}
	if m.TimeLineIndexMemo != nil {
		return setCatAndType(tl, 7, 0)
	}
	if m.TimeLineMonoMemo != nil {
		return setCatAndType(tl, 8, 1)
	}
	if m.TimeLineDoujinMemo != nil {
		return setCatAndType(tl, 9, 0)
	}
	return tl
}

func setCatAndType(tl *TimeLine, cat uint16, typ TimeLineType) *TimeLine {
	tl.Cat = cat
	tl.Type = typ
	return tl
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
	ID        *int
	UserID    *string
	SubjectID *string
	Images    *string
}

type TimeLineImages []TimeLineImage
