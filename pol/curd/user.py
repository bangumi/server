from typing import Optional
from datetime import datetime, timedelta

from loguru import logger
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa
from pol.db.tables import ChiiMember, ChiiOauthAccessToken
from pol.permission import Role, UserGroup
from pol.curd.exceptions import NotFoundError


class User(Role, BaseModel):
    id: int
    username: str
    nickname: str
    group_id: UserGroup
    registration_date: datetime
    sign: str
    avatar: str

    # lastvisit: int
    # lastactivity: int
    # lastpost: int
    # dateformat: str
    # timeformat: int
    # timeoffset: str
    # newpm: int
    # new_notify: int

    def allow_nsfw(self) -> bool:
        allow_date = self.registration_date + timedelta(days=60)
        return datetime.utcnow().astimezone() > allow_date

    def get_username(self) -> str:
        return self.username

    def get_user_id(self) -> int:
        return self.id


async def get_by_valid_token(db: AsyncSession, access_token: str) -> User:
    access: Optional[ChiiOauthAccessToken] = await db.scalar(
        sa.get(
            ChiiOauthAccessToken,
            ChiiOauthAccessToken.access_token == access_token,
            ChiiOauthAccessToken.expires > datetime.now(),
        )
    )

    if not access:
        raise NotFoundError()

    member: ChiiMember = await db.get(ChiiMember, int(access.user_id))

    if not member:
        # 有access token又没有对应的user不太可能发生，如果发生的话打个 log 当作验证失败
        logger.error("can't find user {} for access token", access.user_id)
        raise NotFoundError()

    return User(
        id=member.uid,
        group_id=member.groupid,
        username=member.username,
        nickname=member.nickname,
        registration_date=member.regdate,
        sign=member.sign,
        avatar=member.avatar,
    )
