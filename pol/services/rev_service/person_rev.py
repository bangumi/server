import datetime
from typing import Any, List, Type, Iterator, Optional

from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from pol import curd
from pol.db import sa
from pol.db.const import RevisionType
from pol.db.tables import ChiiRevText, ChiiRevHistory
from pol.services.rev_service.exception import RevisionNotFound

person_rev_type_filters = ChiiRevHistory.rev_type.in_(RevisionType.person_rev_types())


class PersonHistory(BaseModel):
    id: int
    type: int
    created_at: datetime.datetime
    summary: str
    creator_id: int


class PersonHistoryDetail(BaseModel):
    id: int
    type: int
    created_at: datetime.datetime
    summary: str
    creator_id: int

    data: Any


class _PersonRevisionService:
    _db: AsyncSession
    NotFoundError: Type[RevisionNotFound]

    async def get_person_history(self, revision_id) -> PersonHistoryDetail:
        r: Optional[ChiiRevHistory] = await self._db.scalar(
            sa.get(
                ChiiRevHistory,
                ChiiRevHistory.rev_id == revision_id,
                person_rev_type_filters,
            )
        )
        if not r:
            raise self.NotFoundError

        text_item: Optional[ChiiRevText] = await self._db.get(
            ChiiRevText, r.rev_text_id
        )
        if not text_item:
            raise self.NotFoundError

        return PersonHistoryDetail(
            id=r.rev_id,
            type=r.rev_type,
            created_at=r.rev_dateline,
            summary=r.rev_edit_summary,
            data=text_item.rev_text,
            creator_id=r.rev_creator,
        )

    async def count_person_history(self, person_id: int) -> int:
        return await curd.count(
            self._db, person_rev_type_filters, ChiiRevHistory.rev_mid == person_id
        )

    async def list_person_history(
        self, person_id: int, limit: int, offset: int
    ) -> List[PersonHistory]:
        query = (
            sa.select(ChiiRevHistory)
            .where(person_rev_type_filters, ChiiRevHistory.rev_mid == person_id)
            .order_by(ChiiRevHistory.rev_id.desc())
            .limit(limit)
            .offset(offset)
        )

        results: Iterator[ChiiRevHistory] = await self._db.scalars(query)

        return [
            PersonHistory(
                id=r.rev_id,
                type=r.rev_type,
                created_at=r.rev_dateline,
                summary=r.rev_edit_summary,
                creator_id=r.rev_creator,
            )
            for r in results
        ]
