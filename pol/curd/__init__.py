from typing import List, Type, TypeVar, Optional

from databases import Database
from sqlalchemy.orm import DeclarativeMeta

from . import person
from ..db_models import sa
from .exceptions import NotFoundError

T = TypeVar("T", bound=DeclarativeMeta)


async def get_one(db: Database, t: Type[T], *where) -> T:
    query = sa.select(t).where(*where).limit(1)
    r = await db.fetch_one(query)

    if r:
        return t(**r)

    raise NotFoundError()


async def get_one_null(db: Database, t: Type[T], *where) -> Optional[T]:
    query = sa.select(t).where(*where).limit(1)
    r = await db.fetch_one(query)

    if r:
        return t(**r)


async def get_all(db: Database, t: Type[T], *where) -> List[T]:
    query = sa.select(t).where(*where).limit(1)
    result = []
    for r in await db.fetch_all(query):
        result.append(**r)
    return result


__all__ = ["person", "get_one"]
