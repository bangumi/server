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
	*TimeLineMeta
	*TimeLineMemo
}

type TimeLineMeta struct {
	ID  TimeLineID
	UID UserID

	Related  string
	Batch    uint8
	Source   uint8
	Replies  uint32
	Dateline uint32
	Image    TimeLineImages
}

type TimeLineMemo struct {
	Cat  TimeLineCat // Category
	Type uint16

	Content *TimeLineMemoContent
}

type TimeLineMemoContent struct {
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

type TimeLineMemoContentType interface {
	*TimeLineRelationMemo | *TimeLineGroupMemo | *TimeLineWikiMemo | *TimeLineSubjectMemo | *TimeLineProgressMemo |
		*TimeLineSayMemo | *TimeLineBlogMemo | *TimeLineIndexMemo | *TimeLineMonoMemo | *TimeLineDoujinMemo
}

func NewTimeLineMemo[T TimeLineMemoContentType](content T) *TimeLineMemo {
	val := any(content)
	switch typed := val.(type) {
	case *TimeLineRelationMemo:
		return newTimeLineMemo(TimeLineCatRelation, 2, &TimeLineMemoContent{TimeLineRelationMemo: typed})
	case *TimeLineGroupMemo:
		return newTimeLineMemo(TimeLineCatGroup, 3, &TimeLineMemoContent{TimeLineGroupMemo: typed})
	case *TimeLineWikiMemo:
		return newTimeLineMemo(TimeLineCatWiki, 0, &TimeLineMemoContent{TimeLineWikiMemo: typed})
	case *TimeLineSubjectMemo:
		return newTimeLineMemo(TimeLineCatSubject, 0, &TimeLineMemoContent{TimeLineSubjectMemo: typed})
	case *TimeLineProgressMemo:
		return newTimeLineMemo(TimeLineCatProgress, 0, &TimeLineMemoContent{TimeLineProgressMemo: typed})
	case *TimeLineSayMemo:
		if typed.TimeLineSayEdit != nil {
			return newTimeLineMemo(TimeLineCatSay, 2, &TimeLineMemoContent{TimeLineSayMemo: typed})
		}
		return newTimeLineMemo(TimeLineCatSay, 0, &TimeLineMemoContent{TimeLineSayMemo: typed})
	case *TimeLineBlogMemo:
		return newTimeLineMemo(TimeLineCatBlog, 0, &TimeLineMemoContent{TimeLineBlogMemo: typed})
	case *TimeLineIndexMemo:
		return newTimeLineMemo(TimeLineCatIndex, 0, &TimeLineMemoContent{TimeLineIndexMemo: typed})
	case *TimeLineMonoMemo:
		return newTimeLineMemo(TimeLineCatMono, 1, &TimeLineMemoContent{TimeLineMonoMemo: typed})
	case *TimeLineDoujinMemo:
		return newTimeLineMemo(TimeLineCatDoujin, 0, &TimeLineMemoContent{TimeLineDoujinMemo: typed})
	default:
		return nil
	}
}

func newTimeLineMemo(cat TimeLineCat, typ uint16, content *TimeLineMemoContent) *TimeLineMemo {
	return &TimeLineMemo{
		Cat:     cat,
		Type:    typ,
		Content: content,
	}
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
