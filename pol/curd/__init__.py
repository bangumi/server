from typing import Type, TypeVar

from databases import Database

from pol import sa
from . import ep, subject
from ..db.tables import Base
from .exceptions import NotFoundError

T = TypeVar("T", bound=Base)


async def get_one(db: Database, t: Type[T], *where) -> T:
    query = sa.select(t).where(*where).limit(1)
    r = await db.fetch_one(query)

    if r:
        return t(**r)

    raise NotFoundError()


__all__ = ["get_one", "subject", "ep"]
