import enum
from typing import List, Optional

from fastapi import Path, Depends, APIRouter
from databases import Database
from fastapi.exceptions import RequestValidationError
from starlette.responses import Response, RedirectResponse
from sqlalchemy.ext.asyncio import AsyncSession
from pydantic.error_wrappers import ErrorWrapper

from pol import sa, res, wiki
from pol.utils import subject_images
from pol.config import CACHE_KEY_PREFIX
from pol.models import ErrorDetail
from pol.depends import get_db, get_redis, get_session
from pol.db.const import Gender, get_character_rel
from pol.db.tables import ChiiCharacter, ChiiPersonField, ChiiCrtSubjectIndex
from pol.api.v0.const import NotFoundDescription
from pol.api.v0.utils import person_images, short_description
from pol.api.v0.models import (
    Order,
    Paged,
    Pager,
    Character,
    RelatedSubject,
    CharacterDetail,
)
from pol.redis.json_cache import JSONRedis

router = APIRouter(tags=["角色"])

api_base = "/v0/characters"


async def exc_404(character_id: int):
    return res.HTTPException(
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
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_character_detail(
    response: Response,
    db_session: AsyncSession = Depends(get_session),
    not_found: Exception = Depends(exc_404),
    character_id: int = Path(..., gt=0),
    redis: JSONRedis = Depends(get_redis),
):
    cache_key = CACHE_KEY_PREFIX + f"character:{character_id}"
    if value := await redis.get_with_model(cache_key, CharacterDetail):
        response.headers["x-cache-status"] = "hit"
        return value

    character: Optional[ChiiCharacter] = await db_session.scalar(
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

    field = await db_session.get(ChiiPersonField, character_id)
    if field is not None:
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
    db_session: AsyncSession = Depends(get_session),
    not_found: Exception = Depends(exc_404),
    character_id: int = Path(..., gt=0),
):
    character: Optional[ChiiCharacter] = await db_session.scalar(
        sa.select(ChiiCharacter)
        .options(
            sa.selectinload(ChiiCharacter.subjects).joinedload(
                ChiiCrtSubjectIndex.subject
            )
        )
        .where(ChiiCharacter.crt_id == character_id)
        .limit(1)
    )

    if character is None:
        raise not_found

    if character.crt_redirect:
        return RedirectResponse(f"{api_base}/{character.crt_redirect}/subjects")

    if character.crt_ban:
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


def raise_ban(c: ChiiCharacter):
    if c.crt_ban:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description=NotFoundDescription,
            detail={"character_id": "character_id"},
        )
