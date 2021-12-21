from typing import List

import pydantic
from fastapi import Path, Depends, APIRouter
from pydantic import Extra, Field, ValidationError
from databases import Database
from starlette.responses import Response, RedirectResponse

from pol import sa, res, curd, wiki
from pol.config import CACHE_KEY_PREFIX
from pol.models import ErrorDetail
from pol.depends import get_db, get_redis
from pol.db.const import (
    PLATFORM_MAP,
    RELATION_MAP,
    StaffMap,
    SubjectType,
    get_character_rel,
)
from pol.db.tables import (
    ChiiPerson,
    ChiiEpisode,
    ChiiSubject,
    ChiiCharacter,
    ChiiPersonCsIndex,
    ChiiCrtSubjectIndex,
    ChiiSubjectRelations,
)
from pol.permission import Role
from pol.api.v0.const import NotFoundDescription
from pol.api.v0.utils import get_career, person_images, short_description
from pol.api.v0.models import RelPerson, RelCharacter
from pol.curd.exceptions import NotFoundError
from pol.redis.json_cache import JSONRedis
from pol.api.v0.depends.auth import optional_user
from pol.api.v0.models.subject import Subject, RelSubject

router = APIRouter(tags=["条目"])

api_base = "/v0/subjects"


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


@router.get(
    "/subjects/{subject_id}",
    description="cache with 300s",
    response_model_by_alias=False,
    response_model=Subject,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject(
    response: Response,
    db: Database = Depends(get_db),
    subject_id: int = Path(..., gt=0),
    user: Role = Depends(optional_user),
    redis: JSONRedis = Depends(get_redis),
):
    cache_key = CACHE_KEY_PREFIX + f"subject:{subject_id}"
    if value := await redis.get(cache_key):
        try:
            s = Subject.parse_obj(value)
            if s.nsfw:
                if not user.allow_nsfw():
                    raise res.HTTPException(
                        status_code=404,
                        title="Not Found",
                        description=NotFoundDescription,
                        detail={"subject_id": subject_id},
                    )
            response.headers["x-cache-status"] = "hit"
            return value
        except ValidationError:
            await redis.delete(cache_key)
    else:
        response.headers["x-cache-status"] = "miss"

    subject = await basic_subject(db, subject_id)

    if subject.redirect:
        return RedirectResponse(f"{api_base}/{subject.redirect}")

    data = subject.dict(exclude={"dateline", "image", "dropped", "wish"})

    data["locked"] = subject.locked
    data["images"] = subject.images
    data["collection"] = subject.collection
    data["rating"] = subject.rating()
    data["platform"] = PLATFORM_MAP[subject.type].get(
        subject.platform, {"type_cn": ""}
    )["type_cn"]
    data["total_episodes"] = await db.fetch_val(
        sa.select(sa.func.count(ChiiEpisode.ep_id)).where(
            ChiiEpisode.ep_subject_id == subject_id
        )
    )

    try:
        data["infobox"] = wiki.parse(subject.infobox).info
    except wiki.WikiSyntaxError:
        data["infobox"] = None

    if subject.nsfw:
        if not user.allow_nsfw():
            raise res.HTTPException(
                status_code=404,
                title="Not Found",
                description=NotFoundDescription,
                detail={"subject_id": subject_id},
            )

    await redis.set_json(cache_key, value=data, ex=300)

    return data


@router.get(
    "/subjects/{subject_id}/persons",
    response_model_by_alias=False,
    response_model=List[RelPerson],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_persons(
    db: Database = Depends(get_db),
    subject_id: int = Path(..., gt=0),
):
    await basic_subject(db, subject_id, ChiiSubject.subject_ban == 0)

    query = (
        sa.select(
            ChiiPersonCsIndex.prsn_id,
            ChiiPersonCsIndex.prsn_position,
            ChiiPersonCsIndex.subject_type_id,
            ChiiPerson.prsn_name,
            ChiiPerson.prsn_type,
            ChiiPerson.prsn_img,
            ChiiPerson.prsn_summary,
            ChiiPerson.prsn_producer,
            ChiiPerson.prsn_mangaka,
            ChiiPerson.prsn_actor,
            ChiiPerson.prsn_artist,
            ChiiPerson.prsn_seiyu,
            ChiiPerson.prsn_writer,
            ChiiPerson.prsn_illustrator,
        )
        .join(ChiiPerson, ChiiPerson.prsn_id == ChiiPersonCsIndex.prsn_id)
        .where(
            ChiiPersonCsIndex.subject_id == subject_id, ChiiPerson.prsn_redirect == 0
        )
    )

    persons = [
        {
            "id": r["prsn_id"],
            "name": r["prsn_name"],
            "type": r["prsn_type"],
            "relation": StaffMap[r["subject_type_id"]][r["prsn_position"]].str(),
            "career": get_career(r),
            "short_summary": short_description(r["prsn_summary"]),
            "images": person_images(r["prsn_img"]),
        }
        for r in await db.fetch_all(query)
    ]

    return persons


@router.get(
    "/subjects/{subject_id}/characters",
    response_model_by_alias=False,
    response_model=List[RelCharacter],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_characters(
    db: Database = Depends(get_db),
    subject_id: int = Path(..., gt=0),
):
    await basic_subject(db, subject_id, ChiiSubject.subject_ban == 0)

    query = (
        sa.select(
            ChiiCrtSubjectIndex.crt_id,
            ChiiCrtSubjectIndex.crt_type,
            ChiiCharacter.crt_name,
            ChiiCharacter.crt_role,
            ChiiCharacter.crt_img,
            ChiiCharacter.crt_summary,
        )
        .distinct()
        .join(ChiiCharacter, ChiiCharacter.crt_id == ChiiCrtSubjectIndex.crt_id)
        .where(
            ChiiCrtSubjectIndex.subject_id == subject_id,
            ChiiCharacter.crt_redirect == 0,
        )
    )

    characters = [
        {
            "id": r["crt_id"],
            "name": r["crt_name"],
            "relation": get_character_rel(r["crt_type"]),
            "type": r["crt_role"],
            "short_summary": short_description(r["crt_summary"]),
            "images": person_images(r["crt_img"]),
        }
        for r in await db.fetch_all(query)
    ]

    return characters


@router.get(
    "/subjects/{subject_id}/subjects",
    response_model_by_alias=False,
    response_model=List[RelSubject],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_relations(
    db: Database = Depends(get_db),
    subject_id: int = Path(..., gt=0),
):
    subject = await basic_subject(db, subject_id, ChiiSubject.subject_ban == 0)

    query = (
        sa.select(
            ChiiSubjectRelations.rlt_related_subject_type_id,
            ChiiSubjectRelations.rlt_related_subject_id,
            ChiiSubjectRelations.rlt_relation_type,
            ChiiSubject.subject_name,
            ChiiSubject.subject_name_cn,
        )
        .join(
            ChiiSubject,
            ChiiSubject.subject_id == ChiiSubjectRelations.rlt_related_subject_id,
        )
        .where(ChiiSubjectRelations.rlt_subject_id == subject_id)
        .order_by(
            ChiiSubjectRelations.rlt_order, ChiiSubjectRelations.rlt_related_subject_id
        )
    )

    rows = pydantic.parse_obj_as(List[JoinRow], await db.fetch_all(query))

    response = []
    for r in rows:
        relation = RELATION_MAP[r.related_subject_type_id].get(r.relation_type)

        if relation is None or r.relation_type == 1:
            rel = r.related_subject_type_id.str()
        else:
            rel = relation.str()

        response.append(
            {
                "id": r.related_subject_id,
                "relation": rel,
                "name": r.name,
                "type": r.related_subject_type_id,
                "name_cn": r.name_cn,
                "images": subject.images,
            }
        )

    return response


class JoinRow(pydantic.BaseModel):
    """JOIN row result for `get_subject_relations`"""

    related_subject_id: int = Field(alias="rlt_related_subject_id")
    related_subject_type_id: SubjectType = Field(alias="rlt_related_subject_type_id")
    relation_type: int = Field(alias="rlt_relation_type")
    name_cn: str = Field(alias="subject_name_cn")
    name: str = Field(alias="subject_name")

    class Config:
        extra = Extra.forbid
