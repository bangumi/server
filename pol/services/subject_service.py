import datetime
from typing import Dict, List, Iterator, Optional, cast

from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from pol.db import sa
from pol.depends import get_db
from pol.db.const import get_character_rel
from pol.db.tables import ChiiSubject, ChiiCrtSubjectIndex
from pol.models.subject import Subject, SubjectCharacter
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

        results = cast(
            Iterator[ChiiSubject],
            await self._db.scalars(
                sa.select(ChiiSubject)
                .where(*where)
                .options(sa.joinedload(ChiiSubject.fields))
            ),
        )

        return {cast(int, s.subject_id): _convert_from_orm(s) for s in results}

    async def get_by_id(
        self,
        subject_id: int,
        include_nsfw: bool = False,
        include_redirect: bool = False,
        include_banned: bool = False,
    ) -> Subject:
        """

        :param subject_id:
        :param include_nsfw: if nsfw subject should be included
        :param include_redirect: if merged subject included.
        if `include_redirect=True` subject with redirect will raise a `NotFoundError`
        :param include_banned: if `subject_ban` equal to 0.
        """
        where = [ChiiSubject.subject_id == subject_id]
        if not include_banned:
            where.append(ChiiSubject.subject_ban == 0)

        s: Optional[ChiiSubject] = await self._db.scalar(
            sa.get(ChiiSubject, *where).options(sa.joinedload(ChiiSubject.fields)),
        )

        if not s:
            raise self.NotFoundError

        if not include_nsfw and s.subject_nsfw:
            raise self.NotFoundError

        if not include_redirect and s.fields.field_redirect:
            raise self.NotFoundError

        return _convert_from_orm(s)

    async def list_characters(self, id: int):
        s: ChiiSubject = await self._db.scalar(
            sa.select(ChiiSubject).where(
                ChiiSubject.subject_id == id, ChiiSubject.subject_ban == 0
            )
        )

        if not s:
            return []

        res = cast(
            List[ChiiCrtSubjectIndex],
            await self._db.scalars(
                sa.select(ChiiCrtSubjectIndex).where(
                    ChiiCrtSubjectIndex.subject_id == id
                )
            ),
        )

        return [
            SubjectCharacter(
                id=rel.crt_id,
                relation=get_character_rel(rel.crt_type),
            )
            for rel in res
        ]


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
