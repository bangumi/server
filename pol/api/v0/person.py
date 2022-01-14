from typing import List, Optional

from fastapi import Path, Depends, APIRouter
from starlette.responses import Response, RedirectResponse
from sqlalchemy.ext.asyncio import AsyncSession

from pol import res, wiki
from pol.db import sa
from pol.res import ErrorDetail, not_found_exception
from pol.config import CACHE_KEY_PREFIX
from pol.router import ErrorCatchRoute
from pol.depends import get_db, get_redis
from pol.db.const import Gender, StaffMap
from pol.db.tables import (
    ChiiPerson,
    ChiiSubject,
    ChiiCharacter,
    ChiiCrtCastIndex,
    ChiiPersonCsIndex,
    ChiiCharacterField,
)
from pol.api.v0.utils import get_career, person_images, subject_images
from pol.api.v0.models import PersonDetail, RelatedSubject, PersonCharacter
from pol.redis.json_cache import JSONRedis

router = APIRouter(tags=["人物"], route_class=ErrorCatchRoute)

api_base = "/v0/persons"


@router.get(
    "/persons/{person_id}",
    description="cache with 60s",
    response_model=PersonDetail,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person(
    response: Response,
    db: AsyncSession = Depends(get_db),
    person_id: int = Path(..., gt=0),
    not_found: res.HTTPException = Depends(not_found_exception),
    redis: JSONRedis = Depends(get_redis),
):
    cache_key = CACHE_KEY_PREFIX + f"person:{person_id}"
    if value := await redis.get_with_model(cache_key, PersonDetail):
        response.headers["x-cache-status"] = "hit"
        return value.dict()

    person: Optional[ChiiPerson] = await db.scalar(
        sa.select(ChiiPerson).where(ChiiPerson.prsn_id == person_id).limit(1)
    )

    if person is None:
        raise not_found

    if person.prsn_redirect:
        return RedirectResponse(f"{api_base}/{person.prsn_redirect}")

    if person.prsn_ban:
        raise not_found

    data = {
        "id": person.prsn_id,
        "name": person.prsn_name,
        "type": person.prsn_type,
        "career": get_career(person),
        "summary": person.prsn_summary,
        "img": person_img_url(person.prsn_img),
        "images": person_images(person.prsn_img),
        "locked": person.prsn_lock,
        "last_modified": person.prsn_lastpost,
        "stat": {
            "comments": person.prsn_comment,
            "collects": person.prsn_collects,
        },
    }

    field = await db.get(ChiiCharacterField, person_id)
    if field is not None:
        if field.gender:
            data["gender"] = Gender(field.gender).str()
        data["blood_type"] = field.bloodtype or None
        data["birth_year"] = field.birth_year or None
        data["birth_mon"] = field.birth_mon or None
        data["birth_day"] = field.birth_day or None

    try:
        data["infobox"] = wiki.parse(person.prsn_infobox).info
    except wiki.WikiSyntaxError:  # pragma: no cover
        pass

    response.headers["x-cache-status"] = "miss"
    await redis.set_json(cache_key, value=data, ex=60)

    return data


@router.get(
    "/persons/{person_id}/subjects",
    summary="get person related subjects",
    response_model=List[RelatedSubject],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person_subjects(
    db: AsyncSession = Depends(get_db),
    not_found: res.HTTPException = Depends(not_found_exception),
    person_id: int = Path(..., gt=0),
):
    person: ChiiPerson = await db.scalar(
        sa.select(ChiiPerson)
        .options(
            sa.selectinload(ChiiPerson.subjects).joinedload(ChiiPersonCsIndex.subject)
        )
        .where(ChiiPerson.prsn_id == person_id)
        .limit(1)
    )
    if person is None:
        raise not_found

    if person.prsn_redirect:
        return RedirectResponse(f"{api_base}/{person.prsn_redirect}/subjects")

    subjects = []

    for s in person.subjects:
        if v := subject_images(s.subject.subject_image):
            image = v["grid"]
        else:
            image = None

        subjects.append(
            {
                "id": s.subject_id,
                "name": s.subject.subject_name,
                "name_cn": s.subject.subject_name_cn,
                "staff": StaffMap[s.subject_type_id][s.prsn_position].str(),
                "image": image,
            }
        )

    return subjects


@router.get(
    "/persons/{person_id}/characters",
    summary="get person related characters",
    response_model=List[PersonCharacter],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person_characters(
    db: AsyncSession = Depends(get_db),
    not_found: res.HTTPException = Depends(not_found_exception),
    person_id: int = Path(..., gt=0),
):
    person: Optional[ChiiPerson] = await db.scalar(
        sa.select(ChiiPerson)
        .where(ChiiPerson.prsn_id == person_id, ChiiPerson.prsn_ban == 0)
        .limit(1)
    )
    if person is None:
        raise not_found

    query = (
        sa.select(
            ChiiCrtCastIndex.crt_id,
            ChiiCrtCastIndex.prsn_id,
            ChiiCharacter.crt_name,
            ChiiCharacter.crt_role,
            ChiiCharacter.crt_img,
            ChiiSubject.subject_id,
            ChiiSubject.subject_name,
            ChiiSubject.subject_name_cn,
        )
        .distinct()
        .join(ChiiCharacter, ChiiCharacter.crt_id == ChiiCrtCastIndex.crt_id)
        .join(ChiiSubject, ChiiSubject.subject_id == ChiiCrtCastIndex.subject_id)
        .where(
            ChiiCrtCastIndex.prsn_id == person.prsn_id,
            ChiiCharacter.crt_redirect == 0,
            ChiiCharacter.crt_ban == 0,
        )
    )

    characters = [
        {
            "id": r["crt_id"],
            "name": r["crt_name"],
            "type": r["crt_role"],
            "images": person_images(r["crt_img"]),
            "subject_id": r["subject_id"],
            "subject_name": r["subject_name"],
            "subject_name_cn": r["subject_name_cn"],
        }
        for r in (await db.execute(query)).mappings().fetchall()
    ]

    return characters


def person_img_url(s: Optional[str]) -> Optional[str]:
    if not s:
        return None
    return "https://lain.bgm.tv/pic/crt/m/" + s
