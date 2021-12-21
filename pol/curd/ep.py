import datetime
from typing import List

from pydantic import Field, BaseModel
from databases import Database

from pol import sa
from pol.db.tables import ChiiEpisode
from .exceptions import NotFoundError


class Ep(BaseModel):
    id: int = Field(alias="ep_id")
    subject_id: int = Field(alias="ep_subject_id")
    sort: float = Field(alias="ep_sort")
    type: int = Field(alias="ep_type")
    disc: int = Field(0, alias="ep_disc")
    name: str = Field(alias="ep_name")
    name_cn: str = Field(alias="ep_name_cn")
    duration: str = Field(alias="ep_duration")
    airdate: str = Field(alias="ep_airdate")
    online: str = Field(alias="ep_online")
    comment: int = Field(alias="ep_comment")
    # resources: int = Field(alias="ep_resources")
    desc: str = Field(alias="ep_desc")
    dateline: datetime.datetime = Field(alias="ep_dateline")
    lastpost: datetime.datetime = Field(alias="ep_lastpost")
    lock: bool = Field(alias="ep_lock")
    ban: bool = Field(alias="ep_ban")


async def get_many(db: Database, *where, limit=None, offset=None) -> List[Ep]:
    query = (
        sa.select(ChiiEpisode)
        .where(*where)
        .order_by(ChiiEpisode.ep_disc, ChiiEpisode.ep_type, ChiiEpisode.ep_sort)
    )

    if limit is not None:
        query = query.limit(limit)
    if offset is not None:
        query = query.offset(offset)

    results = await db.fetch_all(query)

    return [Ep.parse_obj(r) for r in results]


async def get_one(db: Database, episode_id: int, *where) -> Ep:
    query = sa.select(ChiiEpisode).where(ChiiEpisode.ep_id == episode_id, *where)

    results = await db.fetch_one(query)
    if results:
        return Ep.parse_obj(results)

    raise NotFoundError()
