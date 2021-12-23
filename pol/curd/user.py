from datetime import datetime, timedelta

from loguru import logger
from pydantic import BaseModel
from databases import Database

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


async def get_by_valid_token(db: Database, access_token: str) -> User:
    r = await db.fetch_one(
        sa.get(
            ChiiOauthAccessToken,
            ChiiOauthAccessToken.access_token == access_token,
            ChiiOauthAccessToken.expires > datetime.now(),
        )
    )

    if not r:
        raise NotFoundError()

    access = ChiiOauthAccessToken(**r)

    r = await db.fetch_one(sa.get(ChiiMember, ChiiMember.uid == access.user_id))

    if not r:
        # 有access token又没有对应的user不太可能发生，如果发生的话打个 log 当作验证失败
        logger.error("can't find user {} for access token", access.user_id)
        raise NotFoundError()

    member = ChiiMember(**r)

    return User(
        id=member.uid,
        group_id=member.groupid,
        username=member.username,
        nickname=member.nickname,
        registration_date=member.regdate,
        sign=member.sign,
        avatar=member.avatar,
    )
