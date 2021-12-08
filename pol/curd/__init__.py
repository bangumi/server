from typing import Type, TypeVar

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


__all__ = ["person", "get_one"]
