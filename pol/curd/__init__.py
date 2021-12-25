from typing import Any, Type, TypeVar, Optional

from databases import Database

from pol import sa, res
from pol.db.tables import Base
from pol.api.v0.const import NotFoundDescription
from . import ep, user
from .base import count
from .exceptions import NotFoundError

T = TypeVar("T", bound=Base)


async def get_one(
    db: Database,
    Table: Type[T],
    *where,
    raise_404: bool = False,
    details: Optional[Any] = None,
) -> T:
    query = sa.select(Table).where(*where).limit(1)
    r = await db.fetch_one(query)

    if r:
        t: T = Table(**r)
        return t
    if raise_404:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description=NotFoundDescription,
            detail=details,
        )
    raise NotFoundError()


__all__ = [
    "get_one",
    "count",
    "ep",
    "NotFoundError",
    "user",
]
