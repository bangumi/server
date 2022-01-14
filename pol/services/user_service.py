from typing import Dict, Iterator, Optional
from datetime import datetime

from loguru import logger
from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from pol.db import sa
from pol.models import User, Avatar, PublicUser
from pol.depends import get_db
from pol.db.tables import ChiiMember, ChiiOauthAccessToken
from pol.curd.exceptions import NotFoundError


class UserNotFound(NotFoundError):
    pass


class UserService:
    """user service to get user from database"""

    __slots__ = ("_db",)
    _db: AsyncSession
    NotFoundError = UserNotFound

    @classmethod
    async def new(cls, session: AsyncSession = Depends(get_db)):
        return cls(session)

    def __init__(self, db: AsyncSession):
        self._db = db

    async def get_by_uid(self, uid: int) -> PublicUser:
        """return a public readable user with limited information"""
        u: Optional[ChiiMember] = await self._db.scalar(
            sa.get(ChiiMember, ChiiMember.uid == uid)
        )

        if not u:
            raise self.NotFoundError

        return PublicUser(
            id=u.uid,
            username=u.username,
            nickname=u.nickname,
            avatar=Avatar.from_db_record(u.avatar),
        )

    async def get_users_by_id(self, ids: Iterator[int]) -> Dict[int, PublicUser]:
        """return a public readable user with limited information"""
        results: Iterator[ChiiMember] = await self._db.scalars(
            sa.select(ChiiMember).where(ChiiMember.uid.in_(ids))
        )

        return {
            u.uid: PublicUser(
                id=u.uid,
                username=u.username,
                nickname=u.nickname,
                avatar=Avatar.from_db_record(u.avatar),
            )
            for u in results
        }

    async def get_by_name(self, username: str) -> PublicUser:
        """return a public readable user with limited information"""
        u: Optional[ChiiMember] = await self._db.scalar(
            sa.get(ChiiMember, ChiiMember.username == username)
        )

        if not u:
            raise self.NotFoundError

        return PublicUser(
            id=u.uid,
            username=u.username,
            nickname=u.nickname,
            avatar=Avatar.from_db_record(u.avatar),
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
            raise self.NotFoundError

        member: ChiiMember = await self._db.get(ChiiMember, int(access.user_id))

        if not member:
            # 有access token又没有对应的user不太可能发生，如果发生的话打个 log 当作验证失败
            logger.error(
                "can't find user {user_id} for access token", user_id=access.user_id
            )
            raise self.NotFoundError

        return User(
            id=member.uid,
            group_id=member.groupid,
            username=member.username,
            nickname=member.nickname,
            registration_date=member.regdate,
            sign=member.sign,
            avatar=Avatar.from_db_record(member.avatar),
        )
