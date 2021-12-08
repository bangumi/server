from databases import Database

from pol.db import tables
from pol.db_models import sa
from pol.curd.exceptions import NotFoundError


async def get_by_id(db: Database, id: int) -> tables.ChiiPerson:
    query = sa.select(tables.ChiiPerson).where(tables.ChiiPerson.prsn_id == id).limit(1)
    r = await db.fetch_one(query)

    if r:
        return tables.ChiiPerson(**r)

    raise NotFoundError()
