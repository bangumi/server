from typing import List, Optional

from fastapi import Path, Depends, APIRouter
from starlette.responses import Response
from sqlalchemy.ext.asyncio import AsyncSession

from pol import res, curd, wiki, models
from pol.db import sa
from pol.res import ErrorDetail, not_found_exception
from pol.config import CACHE_KEY_PREFIX
from pol.router import ErrorCatchRoute
from pol.depends import get_db, get_redis
from pol.db.const import (
    PLATFORM_MAP,
    RELATION_MAP,
    StaffMap,
    SubjectType,
    get_character_rel,
)
from pol.db.tables import (
    ChiiEpisode,
    ChiiSubject,
    ChiiPersonCsIndex,
    ChiiCrtSubjectIndex,
    ChiiSubjectRelations,
)
from pol.permission import Role
from pol.api.v0.utils import (
    get_career,
    person_images,
    subject_images,
    short_description,
)
from pol.api.v0.models import RelatedPerson, RelatedCharacter
from pol.api.v0.depends import get_subject
from pol.redis.json_cache import JSONRedis
from pol.http_cache.depends import CacheControl
from pol.api.v0.depends.auth import optional_user
from pol.api.v0.models.subject import Subject, RelatedSubject
from pol.services.subject_service import SubjectService

router = APIRouter(tags=["条目"], route_class=ErrorCatchRoute)

api_base = "/v0/subjects"


@router.get(
    "/subjects/{subject_id}",
    description="cache with 300s",
    summary="获取条目",
    response_model=Subject,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_by_id(
    response: Response,
    not_found: res.HTTPException = Depends(not_found_exception),
    subject_id: int = Path(..., gt=0),
    user: Role = Depends(optional_user),
    db: AsyncSession = Depends(get_db),
    subject_service: SubjectService = Depends(SubjectService.new),
    redis: JSONRedis = Depends(get_redis),
    cache_control: CacheControl = Depends(CacheControl),
):
    """cache and permission wrapper"""
    cache_key = CACHE_KEY_PREFIX + f"res:subject:{subject_id}"
    s = await redis.get_with_model(cache_key, Subject)
    if s:
        response.headers["x-cache-status"] = "hit"
        nsfw = s.nsfw
        data = s.dict()
    else:
        # now fetch real data
        response.headers["x-cache-status"] = "miss"
        data = await _get_subject(
            cache_control,
            not_found=not_found,
            subject_id=subject_id,
            subject_service=subject_service,
            db=db,
        )
        await redis.set_json(cache_key, value=data, ex=300)
        nsfw = data["nsfw"]

    if nsfw and not user.allow_nsfw():
        raise not_found

    return data


async def _get_subject(
    cache_control: CacheControl,
    subject_service: SubjectService,
    not_found: res.HTTPException = Depends(not_found_exception),
    subject_id: int = Path(..., gt=0),
    db: AsyncSession = Depends(get_db),
):
    try:
        s = await subject_service.get_by_id(
            subject_id, include_nsfw=True, include_redirect=True
        )
    except SubjectService.NotFoundError:
        cache_control(300)
        raise not_found

    if s.redirect:
        cache_control(300)
        raise res.HTTPRedirect(f"{api_base}/{s.redirect}")

    # TODO: split these part to another service handler community function
    subject: Optional[ChiiSubject] = await db.get(
        ChiiSubject, subject_id, options=[sa.joinedload(ChiiSubject.fields)]
    )

    if not subject:
        cache_control(300)
        raise not_found

    if not subject.subject_nsfw:
        cache_control(300)

    data = {
        "id": s.id,
        "name": s.name,
        "name_cn": s.name_cn,
        "date": s.date,
        "type": s.type,
        "summary": s.summary,
        "locked": s.locked,
        "nsfw": s.nsfw,
        "images": subject_images(s.image),
        "platform": PLATFORM_MAP[s.type].get(s.platform, {}).get("type_cn", ""),
        "eps": subject.field_eps,
        "volumes": subject.field_volumes,
        "collection": {
            "wish": subject.subject_wish,
            "collect": subject.subject_collect,
            "doing": subject.subject_doing,
            "on_hold": subject.subject_on_hold,
            "dropped": subject.subject_dropped,
        },
        "rating": subject.fields.rating(),
        "total_episodes": await curd.count(db, ChiiEpisode.ep_subject_id == subject_id),
        "tags": subject.fields.tags(),
    }

    try:
        data["infobox"] = wiki.parse(s.infobox).info
    except wiki.WikiSyntaxError:
        data["infobox"] = None

    return data


@router.get(
    "/subjects/{subject_id}/persons",
    response_model=List[RelatedPerson],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_persons(
    db: AsyncSession = Depends(get_db),
    s: models.Subject = Depends(get_subject),
):
    subject: ChiiSubject = await db.scalar(
        sa.select(ChiiSubject)
        .options(
            sa.selectinload(ChiiSubject.persons).joinedload(ChiiPersonCsIndex.person)
        )
        .where(ChiiSubject.subject_id == s.id, ChiiSubject.subject_ban == 0)
    )

    persons = []

    for rel in subject.persons:
        p = rel.person
        persons.append(
            {
                "id": p.prsn_id,
                "name": p.prsn_name,
                "type": p.prsn_type,
                "relation": StaffMap[rel.subject_type_id][rel.prsn_position].str(),
                "career": get_career(p),
                "short_summary": short_description(p.prsn_summary),
                "images": person_images(p.prsn_img),
            }
        )
    return persons


@router.get(
    "/subjects/{subject_id}/characters",
    response_model=List[RelatedCharacter],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_characters(
    db: AsyncSession = Depends(get_db),
    subject: models.Subject = Depends(get_subject),
):
    s: ChiiSubject = await db.scalar(
        sa.select(ChiiSubject)
        .options(
            sa.selectinload(ChiiSubject.characters).joinedload(
                ChiiCrtSubjectIndex.character
            )
        )
        .where(ChiiSubject.subject_id == subject.id, ChiiSubject.subject_ban == 0)
    )

    characters = []
    for rel in s.characters:
        r = rel.character
        characters.append(
            {
                "id": r.crt_id,
                "name": r.crt_name,
                "relation": get_character_rel(rel.crt_type),
                "type": r.crt_role,
                "short_summary": short_description(r.crt_summary),
                "images": person_images(r.crt_img),
            }
        )
    return characters


@router.get(
    "/subjects/{subject_id}/subjects",
    response_model=List[RelatedSubject],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_relations(
    db: AsyncSession = Depends(get_db),
    subject: models.Subject = Depends(get_subject),
):
    s: ChiiSubject = await db.scalar(
        sa.select(ChiiSubject)
        .where(ChiiSubject.subject_id == subject.id, ChiiSubject.subject_ban == 0)
        .options(
            sa.selectinload(ChiiSubject.relations).selectinload(
                ChiiSubjectRelations.dst_subject
            )
        )
    )

    response = []

    for r in s.relations:
        s = r.dst_subject
        relation = RELATION_MAP[r.rlt_related_subject_type_id].get(r.rlt_relation_type)

        if relation is None or r.rlt_related_subject_type_id == 1:
            rel = SubjectType(r.rlt_related_subject_type_id).str()
        else:
            rel = relation.str()

        response.append(
            {
                "id": r.rlt_related_subject_id,
                "relation": rel,
                "name": s.subject_name,
                "type": r.rlt_related_subject_type_id,
                "name_cn": s.subject_name_cn,
                "images": subject_images(s.subject_image),
            }
        )

    return response
