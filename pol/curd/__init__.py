from typing import Type, TypeVar

from databases import Database

from pol import sa
from pol.db.tables import Base
from . import ep, subject
from .exceptions import NotFoundError

T = TypeVar("T", bound=Base)


async def get_one(db: Database, Table: Type[T], *where) -> T:
    query = sa.select(Table).where(*where).limit(1)
    r = await db.fetch_one(query)

    if r:
        t: T = Table(**r)
        return t

    raise NotFoundError()


__all__ = ["get_one", "subject", "ep", "NotFoundError"]
