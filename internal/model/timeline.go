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
	TimeLineCatRelation TimeLineCat = 1
	TimeLineCatGroup    TimeLineCat = 1
	TimeLineCatWiki     TimeLineCat = 2
	TimeLineCatSubject  TimeLineCat = 3
	TimeLineCatProgress TimeLineCat = 4
	TimeLineCatSay      TimeLineCat = 5
	TimeLineCatBlog     TimeLineCat = 6
	TimeLineCatIndex    TimeLineCat = 7
	TimeLineCatMono     TimeLineCat = 8
	TimeLineCatDoujin   TimeLineCat = 9
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

	Cat  TimeLineCat // Category
	Type uint16
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
		return setCatAndType(tl, TimeLineCatRelation, 2)
	}
	if m.TimeLineGroupMemo != nil {
		return setCatAndType(tl, TimeLineCatGroup, 3)
	}
	if m.TimeLineWikiMemo != nil {
		return setCatAndType(tl, TimeLineCatWiki, 0)
	}
	if m.TimeLineSubjectMemo != nil {
		return setCatAndType(tl, TimeLineCatSubject, 0)
	}
	if m.TimeLineProgressMemo != nil {
		return setCatAndType(tl, TimeLineCatProgress, 0)
	}
	if m.TimeLineSayMemo != nil {
		if m.TimeLineSayMemo.TimeLineSayEdit != nil {
			return setCatAndType(tl, TimeLineCatSay, 2)
		}
		return setCatAndType(tl, TimeLineCatSay, 0)
	}
	if m.TimeLineBlogMemo != nil {
		return setCatAndType(tl, TimeLineCatBlog, 0)
	}
	if m.TimeLineIndexMemo != nil {
		return setCatAndType(tl, TimeLineCatIndex, 0)
	}
	if m.TimeLineMonoMemo != nil {
		return setCatAndType(tl, TimeLineCatMono, 1)
	}
	if m.TimeLineDoujinMemo != nil {
		return setCatAndType(tl, TimeLineCatDoujin, 0)
	}
	return tl
}

func setCatAndType(tl *TimeLine, cat uint16, typ TimeLineCat) *TimeLine {
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
