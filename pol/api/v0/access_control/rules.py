from typing import List

from pol.api.v0.access_control.types import ContentPermState, UserPermState, \
    DenialReasonType, ContentRelation, TopicPermState
from pol.db.const import TopicStateType, TopicDisplayStatusType


def can_view_post(state: TopicStateType) -> (bool, DenialReasonType):
    if state == TopicStateType.delete:
        return False, DenialReasonType.deletedByUser
    if state == TopicStateType.private:
        return False, DenialReasonType.guidelineViolation

    return True, DenialReasonType.granted


def can_view_topic(topic_perms: TopicPermState,
                   user_perms: UserPermState,
                   relations: ContentRelation) -> (bool, DenialReasonType):
    #     nsfw_content = ContentPermState.nsfw in content_flags
    # content_closed = ContentPermState.closed in content_flags
    # user_logged_in = UserPermState.exists in user_flags
    # user_banned_from_post = UserPermState.bannedFromPost in user_flags
    # is_content_owner = ContentRelationFlag.contentOwner in relation_flags
    # can_manage_post = UserPermState.canManageTopic in user_flags
    # can_view_silent_post = UserPermState.canViewSilentPost in user_flags
    # can_view_closed_post = UserPermState.canViewClosedPost in user_flags

    # moderator
    if user_perms.canManageTopic:
        return True

    # own content
    if relations.isContentOwner and topic_perms.state in [TopicDisplayStatusType.normal,
                                                          TopicDisplayStatusType.review]:
        return True

    # nsfw
    if topic_perms.nsfw and not user_perms.exists:
        return False


# 审核
if can_manage_post:
    return True

if is_content_owner and

# 下沉
if content_closed and (not user_logged_in or user_banned_from_post or)

# 锁定

# 审核或已被删除
if ContentPermState.hidden not in content_flags:
    return True

return False

#     public static function canViewTopic($topic, $msg = 1) {
#         global $chobits_perm, $chobits_uid;
#         if ($topic) {
#             if ($chobits_perm['manage_topic_state'] ||
#                 ($topic['tpc_uid'] == $chobits_uid) &&
#                  in_array($topic['tpc_display'], array(PostCore::TOPIC_STATUS_NORMAL, PostCore::TOPIC_STATUS_REVIEW))
#                 ) {
#                 return TRUE;
#             }
#             if ($topic['tpc_state'] == Post::TOPIC_STATE_CLOSED && (!$chobits_uid || $chobits_perm['ban_post'] || ($chobits_uid && ValidatorCore::isUserNeedValidate(180)))) {
#                 return FALSE;
#             }
#             if ($topic['tpc_state'] == Post::TOPIC_STATE_SILENT && (!$chobits_uid || $chobits_perm['ban_post'] || ($chobits_uid && ValidatorCore::isUserNeedValidate(365)))) {
#                 return FALSE;
#             }
#             if ($topic['tpc_display'] == PostCore::TOPIC_STATUS_NORMAL) {
#                 return TRUE;
#             }
#         }
#         return FALSE;
#     }
#
#     public static function postContent($post) {
#         if ($post['pst_state'] == Post::TOPIC_STATE_DELETE) {
#             return '<span class="tip">内容已被用户删除</span>';
#         } else if ($post['pst_state'] == Post::TOPIC_STATE_PRIVATE) {
#             return '<span class="tip">内容因违反「<a href="/about/guideline" class="l">社区指导原则</a>」已被删除</span>';
#         }
#         return $post['pst_content'];
#     }
#
#     public static function canViewPost($post) {
#         global $chobits_perm, $chobits_uid;
#         if ($post['pst_state'] == Post::TOPIC_STATE_DELETE || $post['pst_state'] == Post::TOPIC_STATE_PRIVATE) {
#             return FALSE;
#         }
#         return TRUE;
#     }
