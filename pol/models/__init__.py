from datetime import datetime, timedelta

from pydantic import BaseModel

from pol.res import ErrorDetail
from pol.permission import Role, UserGroup

__all__ = ["ErrorDetail", "PublicUser", "User"]


class PublicUser(BaseModel):
    id: int
    username: str
    nickname: str


class User(Role, BaseModel):
    id: int
    username: str
    nickname: str
    group_id: UserGroup
    registration_date: datetime
    sign: str
    avatar: str

    def allow_nsfw(self) -> bool:
        allow_date = self.registration_date + timedelta(days=60)
        return datetime.utcnow().astimezone() > allow_date

    def get_username(self) -> str:
        return self.username

    def get_user_id(self) -> int:
        return self.id
