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

//nolint:tagliatelle
package timeline

import (
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
)

type timelineValue struct {
	Op      string `json:"op"`
	Message any    `json:"message"`
}

type tlEpisode struct {
	ID     model.EpisodeID              `json:"id"`
	Status collection.EpisodeCollection `json:"status"`
}

type tlSubjectCollect struct {
	ID      int               `json:"id"`
	Type    model.SubjectType `json:"type"`
	Eps     int               `json:"eps"`
	Volumes int               `json:"volumes"`
}

type tlCollect struct {
	EpsUpdate  *uint32 `json:"epsUpdate,omitempty"`
	VolsUpdate *uint32 `json:"volsUpdate,omitempty"`
}

type progressSubject struct {
	UID       model.UserID     `json:"uid"`
	Subject   tlSubjectCollect `json:"subject"`
	Collect   tlCollect        `json:"collect"`
	CreatedAt int64            `json:"createdAt"`
	Source    uint8            `json:"source"`
}

type tlCollectRating struct {
	ID      uint64                       `json:"id"`
	Type    collection.SubjectCollection `json:"type"`
	Rate    uint8                        `json:"rate"`
	Comment string                       `json:"comment"`
}

type tlSubject struct {
	ID   model.SubjectID   `json:"id"`
	Type model.SubjectType `json:"type"`
}

type subject struct {
	UID       model.UserID    `json:"uid"`
	Subject   tlSubject       `json:"subject"`
	Collect   tlCollectRating `json:"collect"`
	CreatedAt int64           `json:"createdAt"`
	Source    uint8           `json:"source"`
}

type progressEpisode struct {
	UID       model.UserID `json:"uid"`
	Subject   tlSubject    `json:"subject"`
	Episode   tlEpisode    `json:"episode"`
	CreatedAt int64        `json:"createdAt"`
	Source    uint8        `json:"source"`
}
