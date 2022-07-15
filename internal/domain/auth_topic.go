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

package domain

import "github.com/bangumi/server/internal/model"

func (u Auth) TopicStatuses() []model.TopicStatus {
	if u.ID == 0 {
		return []model.TopicStatus{model.TopicStatusNormal}
	}

	var s = make([]model.TopicStatus, 0, 2)
	s = append(s, model.TopicStatusNormal)

	if u.Permission.ManageTopicState {

	}

	if u.Permission.BanPost {
		s = append(s, model.TopicStatusReview)
	}

	return s
}

/*
public static function canViewTopic($topic, $msg = 1) {
		 global $chobits_perm, $chobits_uid;
		 if ($topic) {
				 if ($chobits_perm['manage_topic_state'] ||
						 ($topic['tpc_uid'] == $chobits_uid) &&
							in_array($topic['tpc_display'], array(PostCore::TOPIC_STATUS_NORMAL, PostCore::TOPIC_STATUS_REVIEW))
						 ) {
						 return TRUE;
				 }
				 if ($topic['tpc_state'] == Post::TOPIC_STATE_CLOSED &&
					(!$chobits_uid || $chobits_perm['ban_post'] || ($chobits_uid && ValidatorCore::isUserNeedValidate(180)))) {
						 return FALSE;
				 }
				 if ($topic['tpc_state'] == Post::TOPIC_STATE_SILENT && (!$chobits_uid || $chobits_perm['ban_post'] || ($chobits_uid && ValidatorCore::isUserNeedValidate(365)))) {
						 return FALSE;
				 }
				 if ($topic['tpc_display'] == PostCore::TOPIC_STATUS_NORMAL) {
						 return TRUE;
				 }
		 }
		 return FALSE;
}
*/
