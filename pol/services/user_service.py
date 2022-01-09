from typing import Optional
from datetime import datetime

from loguru import logger
from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa
from pol.models import User, PublicUser
from pol.depends import get_db
from pol.db.tables import ChiiMember, ChiiOauthAccessToken


class EntityNotFound(Exception):
    pass


class UserService:
    __slots__ = ("_db",)
    _db: AsyncSession
    not_found = EntityNotFound

    @classmethod
    async def new(cls, session: AsyncSession = Depends(get_db)):
        return cls(session)

    def __init__(self, db: AsyncSession):
        self._db = db

    async def get_by_name(self, username: str) -> PublicUser:
        """return a public readable user with limited information"""
        u: Optional[ChiiMember] = await self._db.scalar(
            sa.get(ChiiMember, ChiiMember.username == username)
        )

        if not u:
            raise self.not_found

        return PublicUser(
            id=u.uid,
            username=u.username,
            nickname=u.nickname,
        )

    async def get_by_access_token(self, access_token: str) -> User:
        """return a authorized user"""
        access: Optional[ChiiOauthAccessToken] = await self._db.scalar(
            sa.get(
                ChiiOauthAccessToken,
                ChiiOauthAccessToken.access_token == access_token,
                ChiiOauthAccessToken.expires > datetime.now(),
            )
        )

        if not access:
            raise self.not_found

        member: ChiiMember = await self._db.get(ChiiMember, int(access.user_id))

        if not member:
            # 有access token又没有对应的user不太可能发生，如果发生的话打个 log 当作验证失败
            logger.error(
                "can't find user {user_id} for access token", user_id=access.user_id
            )
            raise EntityNotFound

        return User(
            id=member.uid,
            group_id=member.groupid,
            username=member.username,
            nickname=member.nickname,
            registration_date=member.regdate,
            sign=member.sign,
            avatar=member.avatar,
        )
