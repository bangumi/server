from typing import Dict

from fastapi import Path, Depends, APIRouter, HTTPException
from databases import Database
from starlette.responses import RedirectResponse

from pol import res, curd, wiki
from pol.utils import img_url
from pol.api.v0 import models
from pol.models import ErrorDetail
from pol.depends import get_db
from pol.db.const import StaffMap, get_staff
from pol.db.tables import ChiiPerson, ChiiSubject, ChiiPersonField, ChiiPersonCsIndex
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
        sa.select(
            ChiiPersonCsIndex.subject_id,
            ChiiPersonCsIndex.prsn_position,
            ChiiPersonCsIndex.subject_type_id,
        )
        .where(ChiiPersonCsIndex.prsn_id == person_id)
        .distinct()
        .order_by(ChiiPersonCsIndex.subject_id)
    )

    result: Dict[int, ChiiPersonCsIndex] = {
        r["subject_id"]: ChiiPersonCsIndex(**r) for r in await db.fetch_all(query)
    }

    query = sa.select(
        ChiiSubject.subject_id,
        ChiiSubject.subject_name,
        ChiiSubject.subject_name_cn,
        ChiiSubject.subject_image,
    ).where(ChiiSubject.subject_id.in_(result.keys()))

    subjects = [dict(r) for r in await db.fetch_all(query)]

    for s in subjects:
        s["subject_image"] = img_url(s["subject_image"])
        rel = result[s["subject_id"]]
        s["staff"] = get_staff(StaffMap[rel.subject_type_id][rel.prsn_position])

    data = {
        "id": person.prsn_id,
        "name": person.prsn_name,
        "type": person.prsn_type,
        "infobox": person.prsn_infobox,
        "role": models.PersonRole.from_table(person),
        "summary": person.prsn_summary,
        "img": img_url(person.prsn_img),
        "subjects": subjects,
        "locked": person.prsn_lock,
    }

    try:
        field = await curd.get_one(
            db,
            ChiiPersonField,
            ChiiPersonField.prsn_id == person_id,
            ChiiPersonField.prsn_cat == "prsn",
        )
        data["gender"] = field.gender or None
        data["blood_type"] = field.bloodtype or None
        data["birth_year"] = field.birth_year or None
        data["birth_mon"] = field.birth_mon or None
        data["birth_day"] = field.birth_day or None
    except NotFoundError:
        pass

    try:
        data["wiki"] = wiki.parse(person.prsn_infobox).info
    except wiki.WikiSyntaxError:
        pass

    return data
