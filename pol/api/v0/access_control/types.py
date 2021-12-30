import enum

from pol.db.const import TopicStateType, TopicDisplayStatusType


class ContentPermState:
    nsfw: bool
    displayOverride: bool  # not in [TOPIC_STATUS_ALL, TOPIC_STATUS_BAN])
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
    fallback = 1  # when no authz rule matches
    deletedByUser = 2
    guidelineViolation = 3
    nsfw = 4
    accountAgeUnqualified = 5
