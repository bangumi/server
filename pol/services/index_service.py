import datetime
from typing import Iterator, Optional, cast

from fastapi import Depends
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from pol import curd
from pol.db import sa
from pol.depends import get_db
from pol.db.const import SubjectType
from pol.db.tables import ChiiIndex, ChiiSubject, ChiiIndexRelated
from pol.api.v0.models import Stat
from pol.curd.exceptions import NotFoundError


class IndexNotFound(NotFoundError):
    pass


class Index(BaseModel):
    id: int
    title: str
    desc: str
    stat: Stat
    total: int
    created_at: datetime.datetime
    updated_at: datetime.datetime
    creator_id: int
    ban: bool


class IndexSubject(BaseModel):
    id: int
    type: int
    comment: str
    added_at: datetime.datetime


class IndexService:

    __slots__ = ("_db",)
    _db: AsyncSession
    NotFoundError = IndexNotFound

    @classmethod
    async def new(cls, session: AsyncSession = Depends(get_db)):
        return cls(session)

    def __init__(self, db: AsyncSession):
        self._db = db

    async def get_index_by_id(self, id: int):
        r: Optional[ChiiIndex] = await self._db.scalar(
            sa.get(ChiiIndex, ChiiIndex.idx_id == id, ChiiIndex.idx_ban == 0)
        )
        if not r:
            raise self.NotFoundError
        return Index(
            id=r.idx_id,
            title=r.idx_title,
            desc=r.idx_desc,
            total=r.idx_subject_total,
            stat=Stat(collects=r.idx_collects, comments=r.idx_replies),
            created_at=r.idx_dateline,
            updated_at=r.idx_lasttouch,
            creator_id=r.idx_uid,
            ban=r.idx_ban,
        )

    async def get_index_nsfw_by_id(self, id: int):
        return (
            await self._db.scalar(
                sa.select(ChiiIndexRelated.idx_rlt_id)
                .join(
                    ChiiSubject, ChiiIndexRelated.idx_rlt_sid == ChiiSubject.subject_id
                )
                .where(
                    ChiiIndexRelated.idx_rlt_rid == id, ChiiSubject.subject_nsfw == 1
                )
            )
            is not None
        )

    async def count_index_subjects(
        self, index_id: int, subject_type: Optional[SubjectType]
    ):
        where = [ChiiIndexRelated.idx_rlt_rid == index_id]
        if subject_type:
            where.append(ChiiIndexRelated.idx_rlt_type == subject_type)
        return await curd.count(self._db, *where)

    async def list_index_subjects(
        self,
        index_id: int,
        limit: int,
        offset: int,
        subject_type: Optional[SubjectType],
    ):
        where = [ChiiIndexRelated.idx_rlt_rid == index_id]
        if subject_type:
            where.append(ChiiIndexRelated.idx_rlt_type == subject_type)
        results = cast(
            Iterator[ChiiIndexRelated],
            await self._db.scalars(
                sa.select(ChiiIndexRelated)
                .where(*where)
                .order_by(ChiiIndexRelated.idx_rlt_order)
                .limit(limit)
                .offset(offset)
            ),
        )

        return [
            IndexSubject(
                id=r.idx_rlt_sid,
                type=r.idx_rlt_type,
                comment=r.idx_rlt_comment,
                added_at=r.idx_rlt_dateline,
            )
            for r in results
        ]
