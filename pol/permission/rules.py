from pol.permission.types import UserPermState, TopicPermState, ContentRelation
from pol.permission.types import DenialReasonType
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
    # moderator
    if user_perms.canManageTopic:
        return True, DenialReasonType.granted

    # own content
    if (relations.isContentOwner
        and topic_perms.displayStatus in [
            TopicDisplayStatusType.normal,
            TopicDisplayStatusType.review]):
        return True, DenialReasonType.granted

    # nsfw
    if topic_perms.nsfw and not user_perms.exists:
        return False, DenialReasonType.nsfw

    # 锁定
    if topic_perms.state == TopicStateType.closed and not user_perms.canViewClosedPost:
        return False, DenialReasonType.accountAgeUnqualified

    # 下沉
    if topic_perms.state == TopicStateType.silent and not user_perms.canViewSilentPost:
        return False, DenialReasonType.accountAgeUnqualified

    # visible to public
    if topic_perms.displayStatus == TopicDisplayStatusType.normal:
        return True, DenialReasonType.granted

    return False, DenialReasonType.fallback
