import datetime
from typing import Any, List, Type, Iterator, Optional

from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from pol import curd
from pol.db import sa
from pol.db.const import RevisionType
from pol.db.tables import ChiiRevText, ChiiRevHistory
from pol.services.rev_service.exception import RevisionNotFound


class CharacterHistory(BaseModel):
    id: int
    type: int
    created_at: datetime.datetime
    summary: str
    creator_id: int


class CharacterHistoryDetail(BaseModel):
    id: int
    type: int
    created_at: datetime.datetime
    summary: str
    creator_id: int

    data: Any


character_rev_type_filters = ChiiRevHistory.rev_type.in_(
    RevisionType.character_rev_types()
)


class _CharacterRevisionService:
    _db: AsyncSession
    NotFoundError: Type[RevisionNotFound]

    async def get_character_history(self, revision_id) -> CharacterHistoryDetail:
        r: Optional[ChiiRevHistory] = await self._db.scalar(
            sa.get(
                ChiiRevHistory,
                ChiiRevHistory.rev_id == revision_id,
                character_rev_type_filters,
            )
        )
        if not r:
            raise self.NotFoundError

        text_item: Optional[ChiiRevText] = await self._db.get(
            ChiiRevText, r.rev_text_id
        )
        if not text_item:
            raise self.NotFoundError

        return CharacterHistoryDetail(
            id=r.rev_id,
            type=r.rev_type,
            created_at=r.rev_dateline,
            summary=r.rev_edit_summary,
            data=text_item.rev_text,
            creator_id=r.rev_creator,
        )

    async def count_character_history(self, character_id: int) -> int:
        return await curd.count(
            self._db, character_rev_type_filters, ChiiRevHistory.rev_mid == character_id
        )

    async def list_character_history(
        self, character_id: int, limit: int, offset: int
    ) -> List[CharacterHistory]:
        query = (
            sa.select(ChiiRevHistory)
            .where(character_rev_type_filters, ChiiRevHistory.rev_mid == character_id)
            .order_by(ChiiRevHistory.rev_id.desc())
            .limit(limit)
            .offset(offset)
        )

        results: Iterator[ChiiRevHistory] = await self._db.scalars(query)

        return [
            CharacterHistory(
                id=r.rev_id,
                type=r.rev_type,
                created_at=r.rev_dateline,
                summary=r.rev_edit_summary,
                creator_id=r.rev_creator,
            )
            for r in results
        ]
