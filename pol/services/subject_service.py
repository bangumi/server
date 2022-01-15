import datetime
from typing import Dict, Iterator, Optional

from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from pol.db import sa
from pol.depends import get_db
from pol.db.tables import ChiiSubject
from pol.models.subject import Subject
from pol.curd.exceptions import NotFoundError


class SubjectNotFound(NotFoundError):
    pass


class SubjectService:
    """user service to get user from database"""

    __slots__ = ("_db",)
    _db: AsyncSession
    NotFoundError = SubjectNotFound

    @classmethod
    async def new(cls, session: AsyncSession = Depends(get_db)):
        return cls(session)

    def __init__(self, db: AsyncSession):
        self._db = db

    async def get_by_ids(
        self, *id: int, include_nsfw: bool = False
    ) -> Dict[int, Subject]:
        """

        :param include_nsfw: if nsfw subject should be included
        """
        where = [ChiiSubject.subject_id.in_(id)]
        if not include_nsfw:
            where.append(ChiiSubject.subject_nsfw == 0)

        results: Iterator[ChiiSubject] = await self._db.scalars(
            sa.select(ChiiSubject)
            .where(*where)
            .options(sa.joinedload(ChiiSubject.fields))
        )

        return {s.subject_id: _convert_from_orm(s) for s in results}

    async def get_by_id(
        self,
        subject_id: int,
        include_nsfw: bool = False,
        include_redirect: bool = False,
    ) -> Subject:
        """

        :param subject_id:
        :param include_nsfw: if nsfw subject should be included
        :param include_redirect: if merged subject included.
        if `include_redirect=false` subject with redirect will raise a ``NotFoundError``
        """
        s: Optional[ChiiSubject] = await self._db.get(
            ChiiSubject,
            subject_id,
            options=[sa.joinedload(ChiiSubject.fields)],
        )

        if not s:
            raise self.NotFoundError

        if not include_nsfw and s.subject_nsfw:
            raise self.NotFoundError

        if not include_redirect and s.fields.field_redirect:
            raise self.NotFoundError

        return _convert_from_orm(s)


def _convert_from_orm(s: ChiiSubject) -> Subject:
    """convert ORM model to app model"""
    date = None
    v = s.fields.field_date
    if isinstance(v, datetime.date):
        date = f"{v.year:04d}-{v.month:02d}-{v.day:02d}"

    return Subject(
        id=s.subject_id,
        type=s.subject_type_id,
        name=s.subject_name,
        name_cn=s.subject_name_cn,
        summary=s.field_summary,
        nsfw=bool(s.subject_nsfw),
        date=date,
        platform=s.subject_platform,
        image=s.subject_image,
        infobox=s.field_infobox,
        redirect=s.fields.field_redirect,
        ban=s.subject_ban,
    )
