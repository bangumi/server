from typing import List

from fastapi import Path, Depends, APIRouter, HTTPException
from databases import Database
from starlette.responses import RedirectResponse

from pol import res, curd, wiki
from pol.utils import imgUrl
from pol.api.v0 import models
from pol.models import ErrorDetail
from pol.depends import get_db
from pol.db.tables import ChiiPerson, ChiiPersonField, ChiiPersonCsIndex
from pol.db_models import sa
from pol.curd.exceptions import NotFoundError

router = APIRouter()


@router.get(
    "/person/{person_id}",
    response_model=models.Person,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person(
    person_id: int = Path(..., gt=0),
    db: Database = Depends(get_db),
):
    try:
        person = await curd.get_one(db, ChiiPerson, ChiiPerson.prsn_id == person_id)
    except NotFoundError:
        raise HTTPException(status_code=404, detail="person not found")

    if person.prsn_redirect:
        return RedirectResponse(str(person.prsn_redirect))

    query = (
        sa.select([ChiiPersonCsIndex.subject_id])
        .where(ChiiPersonCsIndex.prsn_id == person_id)
        .distinct()
        .order_by(ChiiPersonCsIndex.subject_id)
    )

    result: List[int] = []
    for r in await db.fetch_all(query):
        result.append(ChiiPersonCsIndex(**r).subject_id)

    m = models.Person(
        id=person.prsn_id,
        name=person.prsn_name,
        type=person.prsn_type,
        infobox=person.prsn_infobox,
        role=person.role(),
        summary=person.prsn_summary,
        img=imgUrl(person.prsn_img),
        subjects=result,
        locked=person.prsn_lock,
    )

    try:
        field = await curd.get_one(
            db,
            ChiiPersonField,
            ChiiPersonField.prsn_id == person_id,
            ChiiPersonField.prsn_cat == "prsn",
        )
        m.gender = field.gender or None
        m.blood_type = field.bloodtype or None
        m.birth_year = field.birth_year or None
        m.birth_mon = field.birth_mon or None
        m.birth_day = field.birth_day or None
    except NotFoundError:
        pass

    try:
        m.wiki = wiki.parse(person.prsn_infobox).info
    except wiki.WikiSyntaxError:
        pass

    return m
