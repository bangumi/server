import enum
from typing import Dict, List, Optional

from fastapi import Path, Depends, APIRouter
from databases import Database
from fastapi.exceptions import RequestValidationError
from starlette.responses import Response, RedirectResponse
from pydantic.error_wrappers import ErrorWrapper

from pol import sa, res, curd, wiki
from pol.utils import subject_images
from pol.config import CACHE_KEY_PREFIX
from pol.models import ErrorDetail
from pol.depends import get_db, get_redis
from pol.db.const import Gender, get_character_rel
from pol.db.tables import (
    ChiiSubject,
    ChiiCharacter,
    ChiiPersonField,
    ChiiCrtSubjectIndex,
)
from pol.api.v0.const import NotFoundDescription
from pol.api.v0.utils import person_images, short_description
from pol.api.v0.models import (
    Order,
    Paged,
    Pager,
    Character,
    SubjectInfo,
    CharacterDetail,
)
from pol.curd.exceptions import NotFoundError
from pol.redis.json_cache import JSONRedis

router = APIRouter(tags=["角色"])

api_base = "/v0/characters"


async def basic_character(
    character_id: int,
    db: Database = Depends(get_db),
) -> ChiiCharacter:
    try:
        return await curd.get_one(
            db,
            ChiiCharacter,
            ChiiCharacter.crt_id == character_id,
        )
    except NotFoundError:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description=NotFoundDescription,
            detail={"character_id": character_id},
        )


class Sort(str, enum.Enum):
    id = "id"
    name = "name"


@router.get(
    "/characters",
    response_model=Paged[Character],
    include_in_schema=False,
)
async def get_characters(
    db: Database = Depends(get_db),
    page: Pager = Depends(),
    name: Optional[str] = None,
    sort: Sort = Sort.id,
    order: Order = Order.desc,
):
    filters = [ChiiCharacter.crt_ban == 0, ChiiCharacter.crt_redirect == 0]
    if name is not None:
        filters.append(ChiiCharacter.crt_name.contains(name))

    count = await db.fetch_val(
        sa.select(sa.func.count(ChiiCharacter.crt_id)).where(*filters)
    )
    if page.offset > count:
        raise RequestValidationError(
            [
                ErrorWrapper(
                    ValueError(f"offset is too big, must be less than {count}"),
                    loc=("query", "offset"),
                )
            ]
        )

    query = (
        sa.select(
            ChiiCharacter.crt_id,
            ChiiCharacter.crt_name,
            ChiiCharacter.crt_role,
            ChiiCharacter.crt_img,
            ChiiCharacter.crt_summary,
            ChiiCharacter.crt_lock,
        )
        .where(*filters)
        .limit(page.limit)
        .offset(page.offset)
    )

    sort_field = ChiiCharacter.crt_id

    if sort == Sort.name:
        sort_field = ChiiCharacter.crt_name

    if order > 0:
        sort_field = sort_field.asc()
    else:
        sort_field = sort_field.desc()

    query = query.order_by(sort_field)

    characters = [
        {
            "id": r["crt_id"],
            "name": r["crt_name"],
            "type": r["crt_role"],
            "short_summary": short_description(r["crt_summary"]),
            "locked": r["crt_lock"],
            "images": person_images(r["crt_img"]),
        }
        for r in await db.fetch_all(query)
    ]

    return {
        "limit": page.limit,
        "offset": page.offset,
        "total": count,
        "data": characters,
    }


@router.get(
    "/characters/{character_id}",
    description="cache with 60s",
    response_model=CharacterDetail,
    response_model_by_alias=False,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_character_detail(
    response: Response,
    db: Database = Depends(get_db),
    character_id: int = Path(..., gt=0),
    redis: JSONRedis = Depends(get_redis),
):
    cache_key = CACHE_KEY_PREFIX + f"character:{character_id}"
    if (value := await redis.get(cache_key)) is not None:
        response.headers["x-cache-status"] = "hit"
        return value
    else:
        response.headers["x-cache-status"] = "miss"

    character: ChiiCharacter = await basic_character(character_id=character_id, db=db)

    if character.crt_redirect:
        return RedirectResponse(f"{api_base}/{character.crt_redirect}")

    raise_ban(character)

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

    try:
        field = await curd.get_one(
            db,
            ChiiPersonField,
            ChiiPersonField.prsn_id == character.crt_id,
            ChiiPersonField.prsn_cat == "crt",
        )
        data["gender"] = Gender(field.gender).str()
        data["blood_type"] = field.bloodtype or None
        data["birth_year"] = field.birth_year or None
        data["birth_mon"] = field.birth_mon or None
        data["birth_day"] = field.birth_day or None
    except NotFoundError:  # pragma: no cover
        pass

    try:
        data["infobox"] = wiki.parse(character.crt_infobox).info
    except wiki.WikiSyntaxError:  # pragma: no cover
        pass

    await redis.set_json(cache_key, value=data, ex=60)

    return data


@router.get(
    "/characters/{character_id}/subjects",
    summary="get character related subjects",
    response_model=List[SubjectInfo],
    response_model_by_alias=False,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person_subjects(
    db: Database = Depends(get_db),
    character_id: int = Path(..., gt=0),
):
    character: ChiiCharacter = await basic_character(character_id=character_id, db=db)
    if character.crt_redirect:
        return RedirectResponse(f"{api_base}/{character.crt_redirect}/subjects")

    raise_ban(character)

    query = (
        sa.select(
            ChiiCrtSubjectIndex.subject_id,
            ChiiCrtSubjectIndex.crt_type,
        )
        .where(ChiiCrtSubjectIndex.crt_id == character.crt_id)
        .distinct()
        .order_by(ChiiCrtSubjectIndex.subject_id)
    )

    result: Dict[int, ChiiCrtSubjectIndex] = {
        r["subject_id"]: ChiiCrtSubjectIndex(**r) for r in await db.fetch_all(query)
    }

    query = sa.select(
        ChiiSubject.subject_id,
        ChiiSubject.subject_name,
        ChiiSubject.subject_name_cn,
        ChiiSubject.subject_image,
    ).where(ChiiSubject.subject_id.in_(result.keys()))

    subjects = [dict(r) for r in await db.fetch_all(query)]

    for s in subjects:
        if v := subject_images(s["subject_image"]):
            s["subject_image"] = v["grid"]
        else:
            s["subject_image"] = None
        rel = result[s["subject_id"]]
        s["staff"] = get_character_rel(rel.crt_type)

    return subjects


def raise_ban(c: ChiiCharacter):
    if c.crt_ban:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description=NotFoundDescription,
            detail={"character_id": "character_id"},
        )
