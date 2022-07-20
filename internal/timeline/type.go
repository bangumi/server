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

package timeline

import "github.com/bangumi/server/internal/model"

type WikiMemo struct {
	SubjectName   string `json:"subject_name"`
	SubjectNameCn string `json:"subject_name_cn"`
	SubjectID     int    `json:"subject_id"`
}

type BlogMemo struct {
	EntryTitle       string `json:"entry_title"`
	EntryDescription string `json:"entry_desc"`
	EntryID          int    `json:"entry_id"`
}

type SayMemo struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type IndexMemo struct {
	IdxID          string `json:"idx_id"`
	IdxTitle       string `json:"idx_title"`
	IdxDescription string `json:"idx_desc"`
}

type MenoMemo struct {
	Name string `json:"name"`
	Cat  int    `json:"cat"`
	ID   int    `json:"id"`
}

type DoujinMemo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
}

type SubjectMemo struct {
	SubjectID      string `json:"subject_id"`
	SubjectTypeID  string `json:"subject_type_id"`
	SubjectName    string `json:"subject_name"`
	SubjectNameCn  string `json:"subject_name_cn"`
	SubjectSeries  string `json:"subject_series"`
	CollectComment string `json:"collect_comment"`
	CollectRate    int    `json:"collect_rate"`
}

type GroupMemo struct {
	GroupID          string `json:"grp_id"`
	GroupName        string `json:"grp_name"`
	GroupTitle       string `json:"grp_title"`
	GroupDescription string `json:"grp_desc"`
}

type RelationMemo struct {
	UserID   string `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type ProgressMemo struct {
	EpName        *string            `json:"ep_name,omitempty" php:"ep_name,omitempty"`
	VolsTotal     *string            `json:"vols_total,omitempty" php:"vols_total,omitempty"`
	SubjectName   *string            `json:"subject_name,omitempty" php:"subject_name,omitempty"`
	EpsUpdate     *int               `json:"eps_update,omitempty" php:"eps_update,omitempty,string"`
	VolsUpdate    *int               `json:"vols_update,omitempty" php:"vols_update,omitempty,string"`
	EpsTotal      *int               `json:"eps_total,omitempty" php:"eps_total,omitempty,string"`
	EpSort        *int               `json:"ep_sort,omitempty" php:"ep_sort,omitempty,string"`
	EpID          *model.EpisodeID   `json:"ep_id,omitempty" php:"ep_id,omitempty,string"`
	SubjectID     *model.SubjectID   `json:"subject_id,omitempty" php:"subject_id,omitempty,string"`
	SubjectTypeID *model.SubjectType `json:"subject_type_id,omitempty" php:"subject_type_id,omitempty,string"`
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

type TimeLineImages []TimeLineImage
