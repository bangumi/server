import datetime
from typing import Any, List, Type, Iterator, Optional

from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from pol import curd
from pol.db import sa
from pol.db.const import RevisionType
from pol.db.tables import ChiiRevHistory, ChiiSubjectRevision
from pol.services.rev_service.exception import RevisionNotFound

subject_rev_type_filters = ChiiRevHistory.rev_type.in_(RevisionType.subject_rev_types())


class SubjectHistory(BaseModel):
    id: int
    type: int
    created_at: datetime.datetime
    summary: str
    creator_id: int


class SubjectHistoryDetail(BaseModel):
    id: int
    type: int
    created_at: datetime.datetime
    summary: str
    creator_id: int

    data: Any


class _SubjectRevisionService:
    _db: AsyncSession
    NotFoundError: Type[RevisionNotFound]

    async def get_subject_history(self, revision_id) -> SubjectHistoryDetail:
        r: Optional[ChiiSubjectRevision] = await self._db.get(
            ChiiSubjectRevision, revision_id
        )
        if not r:
            raise self.NotFoundError

        return SubjectHistoryDetail(
            id=r.rev_id,
            type=r.rev_type,
            created_at=r.rev_dateline,
            summary=r.rev_edit_summary,
            data={
                "subject_id": r.rev_subject_id,
                "name": r.rev_name,
                "name_cn": r.rev_name_cn,
                "vote_field": r.rev_vote_field,
                "type": r.rev_type,
                "type_id": r.rev_type_id,
                "field_infobox": r.rev_field_infobox,
                "field_summary": r.rev_field_summary,
                "field_eps": r.rev_field_eps,
                "platform": r.rev_platform,
            },
            creator_id=r.rev_creator,
        )

    async def count_subject_history(self, subject_id: int) -> int:
        return await curd.count(
            self._db,
            ChiiSubjectRevision.rev_subject_id == subject_id,
        )

    async def list_subject_history(
        self, subject_id: int, limit: int, offset: int
    ) -> List[SubjectHistory]:
        query = (
            sa.select(ChiiSubjectRevision)
            .where(ChiiSubjectRevision.rev_subject_id == subject_id)
            .order_by(ChiiSubjectRevision.rev_dateline.desc())
            .limit(limit)
            .offset(offset)
        )

        results: Iterator[ChiiSubjectRevision] = await self._db.scalars(query)

        return [
            SubjectHistory(
                id=r.rev_id,
                type=r.rev_type,
                created_at=r.rev_dateline,
                summary=r.rev_edit_summary,
                creator_id=r.rev_creator,
            )
            for r in results
        ]
