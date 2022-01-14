from typing import Type, TypeVar, Optional

from sqlalchemy.ext.asyncio import AsyncSession

from pol.db import sa
from pol.db.tables import Base
from .exceptions import NotFoundError

T = TypeVar("T", bound=Base)


async def count(db: AsyncSession, *where) -> int:
    query = sa.select(sa.func.count(1)).where(*where)
    return int(await db.scalar(query))


async def get_one(
    db: AsyncSession,
    Table: Type[T],
    *where,
) -> T:
    query = sa.select(Table).where(*where).limit(1)
    r: Optional[T] = await db.scalar(query)

    if r is not None:
        return r
    raise NotFoundError
