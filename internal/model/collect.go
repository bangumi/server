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
	"time"
)

type CollectPrivacy uint8

const (
	// CollectPrivacyNone 默认公开收藏。
	CollectPrivacyNone CollectPrivacy = 0
	// CollectPrivacySelf 私有收藏，正常计入评分。
	CollectPrivacySelf CollectPrivacy = 1
	// CollectPrivacyBan Shadow Ban, 显示为私有收藏，不计入评分。
	CollectPrivacyBan CollectPrivacy = 2
)

type UserSubjectCollection struct {
	UpdatedAt   time.Time
	Comment     string
	Tags        []string
	VolStatus   uint32
	EpStatus    uint32
	SubjectID   SubjectID
	SubjectType uint8
	Rate        uint8
	Type        SubjectCollection
	Private     bool
}

type UserEpisodeCollection struct {
	ID   EpisodeID
	Type EpisodeCollection
}

type UserSubjectEpisodesCollection map[EpisodeID]UserEpisodeCollection
