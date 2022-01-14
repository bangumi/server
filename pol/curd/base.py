from typing import TypeVar

from sqlalchemy.ext.asyncio import AsyncSession

from pol.db import sa
from pol.db.tables import Base

T = TypeVar("T", bound=Base)


async def count(db: AsyncSession, *where) -> int:
    query = sa.select(sa.func.count(1)).where(*where)
    return int(await db.scalar(query))
