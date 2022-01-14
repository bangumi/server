import datetime
from typing import List, Type, Iterator

from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from pol import curd
from pol.db import sa
from pol.db.const import RevisionType
from pol.db.tables import ChiiEpRevision, ChiiRevHistory
from pol.services.rev_service.exception import RevisionNotFound

episode_rev_type_filters = ChiiRevHistory.rev_type.in_(RevisionType.episode_rev_types())


class EpisodeHistory(BaseModel):
    id: int
    type: int
    created_at: datetime.datetime
    summary: str
    creator_id: int


class EpisodeHistoryDetail(EpisodeHistory):
    episode_ids: str
    infobox: str
    version: str
    subject_id: int


class _EpisodeRevsionService:
    _db: AsyncSession
    NotFoundError: Type[RevisionNotFound]

    async def get_ep_history(self, revision_id) -> EpisodeHistoryDetail:
        r = await self._db.get(ChiiEpRevision, revision_id)
        if not r:
            raise self.NotFoundError

        return EpisodeHistoryDetail(
            id=r.ep_rev_id,
            type=RevisionType.ep,
            created_at=r.rev_dateline,
            summary=r.rev_edit_summary,
            episode_ids=r.rev_eids,
            infobox=r.rev_ep_infobox,
            subject_id=r.rev_sid,
            version=r.rev_version,
            creator_id=r.rev_creator,
        )

    async def count_ep_history(self, ep_id: int) -> int:
        return await curd.count(
            self._db,
            ChiiEpRevision.ep_rev_id,
            ChiiEpRevision.rev_eids.regexp_match(rf"(^|,){ep_id}($|,)"),
        )

    async def list_ep_history(
        self, ep_id: int, limit: int, offset: int
    ) -> List[EpisodeHistory]:
        query = (
            sa.select(ChiiEpRevision)
            .where(ChiiEpRevision.rev_eids.regexp_match(rf"(^|,){ep_id}($|,)"))
            .order_by(ChiiEpRevision.rev_dateline.desc())
            .limit(limit)
            .offset(offset)
        )

        results: Iterator[ChiiEpRevision] = await self._db.scalars(query)

        return [
            EpisodeHistory(
                id=r.ep_rev_id,
                type=RevisionType.ep,
                created_at=r.rev_dateline,
                summary=r.rev_edit_summary,
                creator_id=r.rev_creator,
            )
            for r in results
        ]
