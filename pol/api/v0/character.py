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
from pol.db.const import Gender, get_character_rel
from pol.db.tables import (
    ChiiPerson,
    ChiiSubject,
    ChiiCharacter,
    ChiiPersonField,
    ChiiCrtCastIndex,
    ChiiCrtSubjectIndex,
)
from pol.api.v0.utils import person_images, subject_images
from pol.api.v0.models import RelatedSubject, CharacterDetail, CharacterPerson
from pol.redis.json_cache import JSONRedis

router = APIRouter(tags=["角色"], route_class=ErrorCatchRoute)

api_base = "/v0/characters"


@router.get(
    "/characters/{character_id}",
    description="cache with 60s",
    response_model=CharacterDetail,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_character_detail(
    response: Response,
    db: AsyncSession = Depends(get_db),
    not_found: res.HTTPException = Depends(not_found_exception),
    character_id: int = Path(..., gt=0),
    redis: JSONRedis = Depends(get_redis),
):
    cache_key = CACHE_KEY_PREFIX + f"character:{character_id}"
    if value := await redis.get_with_model(cache_key, CharacterDetail):
        response.headers["x-cache-status"] = "hit"
        return value

    character: Optional[ChiiCharacter] = await db.scalar(
        sa.select(ChiiCharacter).where(ChiiCharacter.crt_id == character_id).limit(1)
    )

    if character is None:
        raise not_found

    if character.crt_redirect:
        return RedirectResponse(f"{api_base}/{character.crt_redirect}")

    if character.crt_ban:
        raise not_found

    data = {
        "id": character.crt_id,
        "name": character.crt_name,
        "type": character.crt_role,
        "summary": character.crt_summary,
        "images": person_images(character.crt_img),
        "locked": character.crt_lock,
        "stat": {
            "comments": character.crt_comment,
            "collects": character.crt_collects,
        },
    }

    field = await db.get(ChiiPersonField, character_id)
    if field is not None:
        if field.gender:
            data["gender"] = Gender(field.gender).str()
        data["blood_type"] = field.bloodtype or None
        data["birth_year"] = field.birth_year or None
        data["birth_mon"] = field.birth_mon or None
        data["birth_day"] = field.birth_day or None

    try:
        data["infobox"] = wiki.parse(character.crt_infobox).info
    except wiki.WikiSyntaxError:  # pragma: no cover
        pass

    response.headers["x-cache-status"] = "miss"
    await redis.set_json(cache_key, value=data, ex=60)

    return data


@router.get(
    "/characters/{character_id}/subjects",
    summary="get character related subjects",
    response_model=List[RelatedSubject],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person_subjects(
    db: AsyncSession = Depends(get_db),
    not_found: res.HTTPException = Depends(not_found_exception),
    character_id: int = Path(..., gt=0),
):
    character: Optional[ChiiCharacter] = await db.scalar(
        sa.select(ChiiCharacter)
        .options(
            sa.selectinload(ChiiCharacter.subjects).joinedload(
                ChiiCrtSubjectIndex.subject
            )
        )
        .where(ChiiCharacter.crt_id == character_id, ChiiCharacter.crt_ban == 0)
        .limit(1)
    )

    if character is None:
        raise not_found

    subjects = []

    for s in character.subjects:
        if v := subject_images(s.subject.subject_image):
            image = v["grid"]
        else:
            image = None

        subjects.append(
            {
                "id": s.subject_id,
                "name": s.subject.subject_name,
                "name_cn": s.subject.subject_name_cn,
                "staff": get_character_rel(s.crt_type),
                "image": image,
            }
        )

    return subjects


@router.get(
    "/characters/{character_id}/persons",
    summary="get character related persons",
    response_model=List[CharacterPerson],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_character_persons(
    db: AsyncSession = Depends(get_db),
    not_found: res.HTTPException = Depends(not_found_exception),
    character_id: int = Path(..., gt=0),
):
    character: Optional[ChiiCharacter] = await db.scalar(
        sa.select(ChiiCharacter)
        .where(ChiiCharacter.crt_id == character_id, ChiiCharacter.crt_ban == 0)
        .limit(1)
    )

    if character is None:
        raise not_found

    query = (
        sa.select(
            ChiiCrtCastIndex.crt_id,
            ChiiCrtCastIndex.prsn_id,
            ChiiPerson.prsn_name,
            ChiiPerson.prsn_type,
            ChiiPerson.prsn_img,
            ChiiSubject.subject_id,
            ChiiSubject.subject_name,
            ChiiSubject.subject_name_cn,
        )
        .distinct()
        .join(ChiiPerson, ChiiPerson.prsn_id == ChiiCrtCastIndex.prsn_id)
        .join(ChiiSubject, ChiiSubject.subject_id == ChiiCrtCastIndex.subject_id)
        .where(
            ChiiCrtCastIndex.crt_id == character.crt_id,
            ChiiPerson.prsn_ban == 0,
        )
    )

    persons = [
        {
            "id": r["prsn_id"],
            "name": r["prsn_name"],
            "type": r["prsn_type"],
            "images": person_images(r["prsn_img"]),
            "subject_id": r["subject_id"],
            "subject_name": r["subject_name"],
            "subject_name_cn": r["subject_name_cn"],
        }
        for r in (await db.execute(query)).mappings().fetchall()
    ]

    return persons
