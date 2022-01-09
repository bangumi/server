import enum
from typing import Optional


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


class Role:
    def allow_nsfw(self) -> bool:
        """if this user can see nsfw contents"""
        raise NotImplementedError()

    def get_user_id(self) -> Optional[int]:
        raise NotImplementedError()
