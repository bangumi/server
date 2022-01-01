from datetime import datetime, timezone
from typing import Optional

from loguru import logger
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa
from pol.curd.exceptions import NotFoundError
from pol.db.tables import ChiiOauthAccessToken, ChiiMember
from pol.permission.roles import Role
from pol.permission.types import UserGroup, UserPermState


class User(BaseModel):
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

    def days_since_registration(self) -> int:
        return (datetime.now(timezone.utc) - self.registration_date).days

    def to_role(self) -> Role:
        days_since_reg = self.days_since_registration()
        can_view_nsfw = days_since_reg > 60
        can_view_closed_post = days_since_reg > 180
        can_view_silent_post = days_since_reg > 365

        perm_state = UserPermState(exists=True,
                                   canViewNsfw=can_view_nsfw,
                                   canViewClosedPost=can_view_closed_post,
                                   canViewSilentPost=can_view_silent_post,
                                   # isBannedFromPost=False,
                                   # todo: load from db
                                   # canManageTopic=False,
                                   # todo: load from role table (chii_usergroup)
                                   )

        return Role(perm_state)


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
