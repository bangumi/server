import enum

from pol.db.const import TopicStateType, TopicDisplayStatusType


class ContentPermState:
    nsfw: bool
    displayOverride: bool  # $topic['tpc_display'] != PostCore::TOPIC_STATUS_NORMAL)
    state: TopicStateType


class TopicPermState(ContentPermState):
    displayStatus: TopicDisplayStatusType


class UserPermState:
    exists: bool
    canManageTopic: bool
    isBannedFromPost: bool
    canViewSilentPost: bool  # https://bangumi.tv/group/topic/359603
    canViewClosedPost: bool  # https://bangumi.tv/group/topic/359603


class ContentRelation:
    isContentOwner: bool


class DenialReasonType(enum.IntEnum):
    granted = 0
    deletedByUser = 1
    guidelineViolation = 2
