import enum
from datetime import datetime, timedelta

from pydantic import BaseModel


class Avatar(BaseModel):
    large: str
    medium: str
    small: str

    @classmethod
    def from_db_record(cls, s: str):
        """default user user avatar https://lain.bgm.tv/pic/user/l/"""
        if not s:
            s = "icon.jpg"
        return cls(
            large="https://lain.bgm.tv/pic/user/l/" + s,
            medium="https://lain.bgm.tv/pic/user/m/" + s,
            small="https://lain.bgm.tv/pic/user/s/" + s,
        )


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


class Permission(BaseModel):
    """从当前的 chii_usergroup 表中导出的旧主站使用的实际权限列表"""

    app_erase: int = 0
    manage_app: int = 0
    user_list: int = 0
    manage_user_group: int = 0
    manage_user_photo: int = 0
    manage_topic_state: int = 0
    manage_report: int = 0
    user_ban: int = 0
    manage_user: int = 0
    user_group: UserGroup = UserGroup.normal_user
    user_wiki_approve: int = 0
    subject_edit: int = 0
    subject_lock: int = 0
    subject_refresh: int = 0
    subject_related: int = 0
    subject_merge: int = 0
    subject_erase: int = 0
    subject_cover_lock: int = 0
    subject_cover_erase: int = 0
    mono_edit: int = 0
    mono_lock: int = 0
    mono_merge: int = 0
    mono_erase: int = 0
    ep_edit: int = 0
    ep_move: int = 0
    ep_merge: int = 0
    ep_lock: int = 0
    ep_erase: int = 0
    report: int = 0
    doujin_subject_erase: int = 0
    doujin_subject_lock: int = 0


class PublicUser(BaseModel):
    id: int
    username: str
    nickname: str
    avatar: Avatar


class User(BaseModel):
    """private authorized user"""

    id: int
    username: str
    nickname: str
    group_id: UserGroup
    registration_date: datetime
    sign: str
    avatar: Avatar
    permission: Permission

    def allow_nsfw(self) -> bool:
        allow_date = self.registration_date + timedelta(days=60)
        return datetime.utcnow().astimezone() > allow_date

    def get_user_id(self) -> int:
        return self.id
