from typing import List, Optional

from fastapi import Path, Depends, APIRouter
from starlette.responses import Response, RedirectResponse
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa, res, wiki
from pol.utils import subject_images
from pol.config import CACHE_KEY_PREFIX
from pol.models import ErrorDetail
from pol.router import ErrorLoggingRoute
from pol.depends import get_db, get_redis
from pol.db.const import Gender, get_character_rel
from pol.db.tables import ChiiCharacter, ChiiPersonField, ChiiCrtSubjectIndex
from pol.api.v0.const import NotFoundDescription
from pol.api.v0.utils import person_images
from pol.api.v0.models import RelatedSubject, CharacterDetail
from pol.redis.json_cache import JSONRedis

router = APIRouter(tags=["角色"], route_class=ErrorLoggingRoute)

api_base = "/v0/characters"


async def exc_404(character_id: int):
    return res.HTTPException(
        status_code=404,
        title="Not Found",
        description=NotFoundDescription,
        detail={"character_id": character_id},
    )


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
    not_found: Exception = Depends(exc_404),
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
    not_found: Exception = Depends(exc_404),
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
