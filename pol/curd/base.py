from typing import List, Type, TypeVar

from databases import Database

from pol import sa
from pol.db.tables import Base

T = TypeVar("T", bound=Base)


async def count(db: Database, column, *where) -> int:
    query = sa.select(sa.func.count(column)).where(*where)
    return int(await db.fetch_val(query))


async def get_many(
    db: Database,
    Table: Type[T],
    *where,
    order=None,
    limit=None,
    offset=None,
) -> List[T]:
    query = sa.select(Table)

    if where:
        query = query.where(*where)
    if order is not None:
        query = query.order_by(order)
    if limit is not None:
        query = query.limit(limit)
    if offset is not None:
        query = query.offset(offset)

    results = await db.fetch_all(query)

    return [Table(**r) for r in results]
