import enum

from pydantic.main import BaseModel

from pol.db.const import TopicStateType, TopicDisplayStatusType


class UserGroup(enum.IntEnum):
    admin = 1  # 管理员
    bangumi_admin = 2  # Bangumi 管理猿
    window_admin = 3  # 天窗管理猿
    quite_user = 4  # 禁言用户
    banned_user = 5  # 禁止访问用户
    character_admin = 8  # 人物管理猿
    wiki_admin = 9  # 维基条目管理猿
    normal_user = 10  # 用户
    wiki = 11  # 维基人


class ContentPermState(BaseModel):
    nsfw: bool = False
    state: TopicStateType


class TopicPermState(ContentPermState):
    displayStatus: TopicDisplayStatusType


class UserPermState(BaseModel):
    exists: bool = False
    canManageTopic: bool = False
    isBannedFromPost: bool = False
    canViewNsfw: bool = False
    canViewSilentPost: bool = False  # https://bangumi.tv/group/topic/359603
    canViewClosedPost: bool = False  # https://bangumi.tv/group/topic/359603


class ContentRelation:
    isContentOwner: bool = False


class DenialReasonType(enum.IntEnum):
    granted = 0
    fallback = 1  # when no authz rule matches
    deletedByUser = 2
    guidelineViolation = 3
    nsfw = 4
    accountAgeUnqualified = 5
