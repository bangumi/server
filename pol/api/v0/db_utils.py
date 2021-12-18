from databases import Database

from pol import res, curd
from pol.curd import NotFoundError
from pol.db.tables import ChiiSubject
from pol.api.v0.const import NotFoundDescription


async def basic_subject(db: Database, subject_id: int, *where) -> curd.subject.Subject:
    try:
        return await curd.subject.get_one(
            db,
            ChiiSubject.subject_id == subject_id,
            *where,
        )
    except NotFoundError:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description=NotFoundDescription,
            detail={"subject_id": subject_id},
        )
