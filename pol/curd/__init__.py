from typing import Any, Type, TypeVar, Optional

from databases import Database

from pol import sa
from pol.db.tables import Base
from . import ep, user
from .base import count
from .exceptions import NotFoundError

T = TypeVar("T", bound=Base)


async def get_one(
    db: Database,
    Table: Type[T],
    *where,
    details: Optional[Any] = None,
) -> T:
    query = sa.select(Table).where(*where).limit(1)
    r = await db.fetch_one(query)

    if r:
        t: T = Table(**r)
        return t
    raise NotFoundError(details)


__all__ = [
    "get_one",
    "count",
    "ep",
    "NotFoundError",
    "user",
]
