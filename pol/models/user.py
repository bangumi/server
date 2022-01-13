from datetime import datetime, timedelta

from pydantic import BaseModel

from pol.permission import Role, UserGroup


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


class PublicUser(BaseModel):
    id: int
    username: str
    nickname: str
    avatar: Avatar


class User(Role, BaseModel):
    """private authorized user"""

    id: int
    username: str
    nickname: str
    group_id: UserGroup
    registration_date: datetime
    sign: str
    avatar: Avatar

    def allow_nsfw(self) -> bool:
        allow_date = self.registration_date + timedelta(days=60)
        return datetime.utcnow().astimezone() > allow_date

    def get_user_id(self) -> int:
        return self.id
