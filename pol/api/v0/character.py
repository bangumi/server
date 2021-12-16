import enum
from typing import Dict, List, Optional

import pydantic
from fastapi import Path, Query, Depends, APIRouter
from pydantic import Field
from databases import Database
from fastapi.exceptions import RequestValidationError
from starlette.responses import Response, RedirectResponse
from pydantic.error_wrappers import ErrorWrapper

from pol import res, curd, wiki
from pol.utils import subject_images
from pol.api.v0 import models
from pol.config import CACHE_KEY_PREFIX
from pol.models import ErrorDetail
from pol.depends import get_db, get_redis
from pol.db.const import Gender, StaffMap, PersonType, get_character_rel
from pol.db.tables import ChiiSubject, ChiiCharacter, ChiiPersonField, ChiiPersonCsIndex
from pol.db_models import sa
from pol.api.v0.models import PersonCareer
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
            ChiiCharacter.prsn_id == character_id,
            ChiiCharacter.prsn_ban != 1,
        )
    except NotFoundError:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description="resource you resource can't be found in the database",
            detail={"character_id": "character_id"},
        )


class Pager(pydantic.BaseModel):
    limit: int = Field(30, gt=0, le=50)
    offset: int = Field(0, ge=0)


class Sort(str, enum.Enum):
    id = "id"
    name = "name"
    last_modified = "update"


class Order(enum.IntEnum):
    asc = 1
    desc = -1


@router.get(
    "/characters",
    response_model=models.PagedPerson,
)
async def get_persons(
    db: Database = Depends(get_db),
    page: Pager = Depends(),
    name: Optional[str] = None,
    type: Optional[PersonType] = Query(None, description="`1`为个人，`2`为公司，`3`为组合"),
    career: Optional[List[PersonCareer]] = Query(
        None, example="?career=mangaka&career=producer"
    ),
    sort: Sort = Sort.id,
    order: Order = Order.desc,
):
    filters = [ChiiCharacter.prsn_ban == 0, ChiiCharacter.prsn_redirect == 0]
    if name is not None:
        filters.append(ChiiCharacter.prsn_name.contains(name))
    if type is not None:
        filters.append(ChiiCharacter.prsn_type == type.value)

    if career:
        q = []
        for c in career:
            q.append(getattr(ChiiCharacter, f"prsn_{c}") == 1)
        career_filter = sa.or_(*q)
        filters.append(career_filter)

    count = await db.fetch_val(
        sa.select(sa.func.count(ChiiCharacter.prsn_id)).where(*filters)
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
            ChiiCharacter.prsn_id,
            ChiiCharacter.prsn_name,
            ChiiCharacter.prsn_type,
            ChiiCharacter.prsn_img,
            ChiiCharacter.prsn_summary,
            ChiiCharacter.prsn_producer,
            ChiiCharacter.prsn_mangaka,
            ChiiCharacter.prsn_actor,
            ChiiCharacter.prsn_lock,
            ChiiCharacter.prsn_artist,
            ChiiCharacter.prsn_seiyu,
            ChiiCharacter.prsn_writer,
            ChiiCharacter.prsn_illustrator,
        )
        .where(*filters)
        .limit(page.limit)
        .offset(page.offset)
    )

    sort_field = ChiiCharacter.prsn_id

    if sort == Sort.name:
        sort_field = ChiiCharacter.prsn_name
    if sort == Sort.last_modified:
        sort_field = ChiiCharacter.prsn_lastpost

    if order > 0:
        sort_field = sort_field.asc()
    else:
        sort_field = sort_field.desc()

    query = query.order_by(sort_field)

    characters = [
        {
            "id": r["prsn_id"],
            "name": r["prsn_name"],
            "type": r["prsn_type"],
            "career": get_career(r),
            "short_summary": r["prsn_summary"][:80] + "...",
            "locked": r["prsn_lock"],
            "img": person_img_url(r["prsn_img"]),
            "images": person_images(r["prsn_img"]),
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
    response_model=models.PersonDetail,
    response_model_by_alias=False,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person(
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

    if character.prsn_redirect:
        return RedirectResponse(f"{api_base}/{character.prsn_redirect}")

    data = {
        "id": character.prsn_id,
        "name": character.prsn_name,
        "type": character.prsn_type,
        "career": get_career(character),
        "summary": character.prsn_summary,
        "img": person_img_url(character.prsn_img),
        "images": person_images(character.prsn_img),
        "locked": character.prsn_lock,
        "last_modified": character.prsn_lastpost,
        "stat": {
            "comments": character.prsn_comment,
            "collects": character.prsn_collects,
        },
    }

    try:
        field = await curd.get_one(
            db,
            ChiiPersonField,
            ChiiPersonField.prsn_id == character.prsn_id,
            ChiiPersonField.prsn_cat == "prsn",
        )
        data["gender"] = Gender.to_view(field.gender)
        data["blood_type"] = field.bloodtype or None
        data["birth_year"] = field.birth_year or None
        data["birth_mon"] = field.birth_mon or None
        data["birth_day"] = field.birth_day or None
    except NotFoundError:
        pass

    try:
        data["infobox"] = wiki.parse(character.prsn_infobox).info
    except wiki.WikiSyntaxError:
        pass

    await redis.set_json(cache_key, value=data, ex=60)

    return data


@router.get(
    "/characters/{character_id}/subjects",
    summary="get character related subjects",
    response_model=List[models.SubjectInfo],
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
    if character.prsn_redirect:
        return RedirectResponse(f"{api_base}/{character.prsn_redirect}/subjects")

    query = (
        sa.select(
            ChiiPersonCsIndex.subject_id,
            ChiiPersonCsIndex.prsn_position,
            ChiiPersonCsIndex.subject_type_id,
        )
        .where(ChiiPersonCsIndex.prsn_id == character.prsn_id)
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
        if v := subject_images(s["subject_image"]):
            s["subject_image"] = v["grid"]
        else:
            s["subject_image"] = None
        rel = result[s["subject_id"]]
        s["staff"] = get_character_rel(StaffMap[rel.subject_type_id][rel.prsn_position])

    return subjects


def get_career(p: ChiiCharacter) -> List[str]:
    s = []
    if p.prsn_producer:
        s.append("producer")
    if p.prsn_mangaka:
        s.append("mangaka")
    if p.prsn_artist:
        s.append("artist")
    if p.prsn_seiyu:
        s.append("seiyu")
    if p.prsn_writer:
        s.append("writer")
    if p.prsn_illustrator:
        s.append("illustrator")
    if p.prsn_actor:
        s.append("actor")
    return s


def person_images(s: Optional[str]) -> Optional[Dict[str, str]]:
    if not s:
        return None

    return {
        "large": "https://lain.bgm.tv/pic/crt/l/" + s,
        "medium": "https://lain.bgm.tv/pic/crt/m/" + s,
        "small": "https://lain.bgm.tv/pic/crt/s/" + s,
        "grid": "https://lain.bgm.tv/pic/crt/g/" + s,
    }


def person_img_url(s: Optional[str]) -> Optional[str]:
    if not s:
        return None
    return "https://lain.bgm.tv/pic/crt/m/" + s
