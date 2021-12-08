from typing import Dict, List

from fastapi import Path, Depends, Request, APIRouter
from databases import Database
from starlette.responses import RedirectResponse

from pol import res, curd, wiki
from pol.utils import person_img_url, subject_img_url
from pol.api.v0 import models
from pol.models import ErrorDetail
from pol.depends import get_db
from pol.db.const import Gender, StaffMap, BloodType, get_staff
from pol.db.tables import ChiiPerson, ChiiSubject, ChiiPersonField, ChiiPersonCsIndex
from pol.db_models import sa
from pol.curd.exceptions import NotFoundError

router = APIRouter()

api_base = "/api/v0/persons"


async def basic_person(
    request: Request,
    person_id: int = Path(..., gt=0),
    db: Database = Depends(get_db),
) -> ChiiPerson:
    try:
        return await curd.get_one(db, ChiiPerson, ChiiPerson.prsn_id == person_id)
    except NotFoundError:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description="resource you resource can't be found in the database",
            detail={"person_id": request.path_params.get("person_id")},
        )


@router.get(
    "/persons/{person_id}",
    response_model=models.Person,
    response_model_by_alias=False,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person(
    db: Database = Depends(get_db),
    person: ChiiPerson = Depends(basic_person),
):
    if person.prsn_redirect:
        return RedirectResponse(f"{api_base}/{person.prsn_redirect}")

    data = {
        "id": person.prsn_id,
        "name": person.prsn_name,
        "type": person.prsn_type,
        "infobox": person.prsn_infobox,
        "role": [
            key
            for key, value in models.PersonRole.from_orm(person).dict().items()
            if value
        ],
        "summary": person.prsn_summary,
        "img": person_img_url(person.prsn_img),
        "locked": person.prsn_lock,
        "last_update": person.prsn_lastpost,
        "stat": {
            "comments": person.prsn_comment,
            "collects": person.prsn_collects,
        },
    }

    try:
        field = await curd.get_one(
            db,
            ChiiPersonField,
            ChiiPersonField.prsn_id == person.prsn_id,
            ChiiPersonField.prsn_cat == "prsn",
        )
        data["gender"] = Gender.to_view(field.gender)
        data["blood_type"] = BloodType.to_view(field.bloodtype)
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


@router.get(
    "/persons/{person_id}/subjects",
    response_model=List[models.SubjectInfo],
    response_model_by_alias=False,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person_subjects(
    db: Database = Depends(get_db),
    person: ChiiPerson = Depends(basic_person),
):
    if person.prsn_redirect:
        return RedirectResponse(f"{api_base}/{person.prsn_redirect}/subjects")

    query = (
        sa.select(
            ChiiPersonCsIndex.subject_id,
            ChiiPersonCsIndex.prsn_position,
            ChiiPersonCsIndex.subject_type_id,
        )
        .where(ChiiPersonCsIndex.prsn_id == person.prsn_id)
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
        s["subject_image"] = subject_img_url(s["subject_image"])
        rel = result[s["subject_id"]]
        s["staff"] = get_staff(StaffMap[rel.subject_type_id][rel.prsn_position])

    return subjects
